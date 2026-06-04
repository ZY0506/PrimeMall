package orderlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type OrderItemSnapshot struct {
	SkuId       uint64 `json:"sku_id"`
	SpuId       uint64 `json:"spu_id"`
	SpuName     string `json:"spu_name"`
	SkuName     string `json:"sku_name"`
	SkuPic      string `json:"sku_pic"`
	Price       int64  `json:"price"`
	Count       int64  `json:"count"`
	TotalAmount int64  `json:"total_amount"`
}

type SettlementData struct {
	UserId          uint64              `json:"user_id"`
	AddressSnapshot string              `json:"address_snapshot"`
	CouponId        uint64              `json:"coupon_id"`
	ItemSnapshots   []OrderItemSnapshot `json:"item_snapshots"`
	TotalAmount     int64               `json:"total_amount"`
	FreightAmount   int64               `json:"freight_amount"`
	CouponDiscount  int64               `json:"coupon_discount"`
	PayAmount       int64               `json:"pay_amount"`
}

func NewPreOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreOrderLogic {
	return &PreOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PreOrder 预下单
// 核心职责：校验参数/用户/商品/地址 → 计算价格/运费 → 生成订单项快照 → 缓存结算上下文
func (l *PreOrderLogic) PreOrder(in *order.PreOrderRequest) (*order.PreOrderResponse, error) {
	// ===================== 1. 基础参数校验 =====================
	for _, item := range in.Items {
		if item.SkuId <= 0 || item.Quantity <= 0 {
			return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "商品参数错误")
		}
	}
	if in.AddressId <= 0 {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "收货地址不能为空")
	}

	// ===================== 2. 获取登录用户ID =====================
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败，error=%v", err)
		return nil, err
	}

	// ===================== 3. 校验用户下单权限 =====================
	ret, err := l.svcCtx.UserRpc.CheckUserStatus(l.ctx, &user.CheckUserStatusReq{
		UserId:    userId,
		CheckType: constants.USER_STATUS_RESTRICTED,
	})
	if err != nil {
		l.Logger.Errorf("校验用户状态失败，error=%v", err)
		return nil, err
	}
	if !ret.Allowed {
		return nil, errorx.NewBizError(response.ErrCodeUserRestricted, ret.Reason)
	}

	// ===================== 4. 查询收货地址 + 生成地址快照 =====================
	addr := &user.AddressItem{}
	addrSnapshot := new(order.AddressSnapshot)
	var addrSnapshotStr []byte
	if in.AddressId > 0 {
		addr, err = l.svcCtx.UserRpc.GetAddressById(l.ctx, &user.GetAddressReq{
			AddressId: in.AddressId,
			UserId:    userId,
		})
		if err != nil {
			l.Logger.Errorf("查询地址失败，error=%v", err)
			return nil, err
		}
		addrSnapshot = &order.AddressSnapshot{
			ReceiverName:  addr.ReceiverName,
			ReceiverPhone: addr.ReceiverPhone,
			Detail: &order.AddressDetail{
				Province:      addr.Address.Province,
				City:          addr.Address.City,
				District:      addr.Address.District,
				DetailAddress: addr.Address.Detail,
				PostalCode:    addr.Address.PostalCode,
			},
		}
		// 序列化地址快照
		addrSnapshotStr, err = json.Marshal(addrSnapshot)
		if err != nil {
			l.Logger.Errorf("序列化地址快照失败，error=%v", err)
			return nil, err
		}
	}

	// TODO: 获取优惠券信息(校验优惠卷)

	// ===================== 5. 构建商品参数 =====================
	skuIds := make([]uint64, len(in.Items))
	stockCheckItems := make([]*product.SkuStockItem, 0, len(in.Items))
	skuQuantityMap := make(map[uint64]int64, len(in.Items))
	for i, item := range in.Items {
		skuIds[i] = item.SkuId
		skuQuantityMap[item.SkuId] = item.Quantity
		stockCheckItems = append(stockCheckItems, &product.SkuStockItem{
			SkuId:    item.SkuId,
			Quantity: item.Quantity,
		})
	}

	// ===================== 6. 查询商品SKU信息 =====================
	skus, err := l.svcCtx.ProductRpc.GetSkuListByIds(l.ctx, &product.SkuIdsReq{SkuIds: skuIds})
	if err != nil {
		l.Logger.Errorf("获取SKU列表失败，error=%v", err)
		return nil, err
	}
	if len(skus.SkuItems) == 0 {
		return nil, errorx.NewBizError(response.ErrCodePreOrderFailed, "商品不存在")
	}

	// ===================== 7. 检查库存（仅检查，不锁定） =====================
	var stockErrMsgs []string
	var itemSnapshots []OrderItemSnapshot

	for _, sku := range skus.SkuItems {
		if sku.Status != constants.PRODUCT_SKU_STATUS_ENABLED {
			return nil, errorx.NewBizError(response.ErrCodeProductOffline, "商品已下架")
		}
		buyNum := skuQuantityMap[sku.Id]
		// 库存校验（使用可用库存：总库存 - 锁定库存）
		availableStock := sku.Stock - sku.LockedStock
		if availableStock < buyNum {
			stockErrMsgs = append(stockErrMsgs, fmt.Sprintf("商品【%s】库存不足，当前可用库存：%d，购买数量：%d", sku.SpuName, availableStock, buyNum))
		}
		// 拼接规格
		var specStrs []string
		for _, spec := range sku.Specs {
			specStrs = append(specStrs, spec.Value)
		}
		skuName := strings.Join(specStrs, ",")
		pic := ""
		if len(sku.Images) > 0 {
			pic = sku.Images[0]
		}
		// 构建快照
		itemSnapshots = append(itemSnapshots, OrderItemSnapshot{
			SkuId:       sku.Id,
			SpuId:       sku.SpuId,
			SpuName:     sku.SpuName,
			SkuName:     skuName,
			SkuPic:      pic,
			Price:       sku.Price,
			Count:       buyNum,
			TotalAmount: sku.Price * buyNum,
		})
	}
	if len(stockErrMsgs) > 0 {
		errMsg := strings.Join(stockErrMsgs, "；")
		l.Logger.Errorf("商品库存不足：%s", errMsg)
		return nil, errorx.NewBizError(response.ErrCodePreOrderFailed, errMsg)
	}

	// ===================== 8. 计算价格 =====================
	var totalAmount int64 = 0
	for _, item := range itemSnapshots {
		totalAmount += item.TotalAmount
	}

	// ===================== 9. 计算运费 =====================
	var freightAmount int64 = 0
	freightResp, err := l.svcCtx.ProductRpc.CalculateFreight(l.ctx, &product.CalculateFreightReq{
		AddressId: in.AddressId,
		Items:     stockCheckItems,
	})
	if err != nil {
		l.Logger.Errorf("计算运费失败，error=%v", err)
		return nil, err
	}
	freightAmount = freightResp.TotalPrice

	// ===================== 10. 优惠计算 =====================
	var couponDiscount int64 = 0
	// TODO: 计算优惠卷
	payAmount := totalAmount + freightAmount - couponDiscount
	if payAmount < 0 {
		payAmount = 0
	}

	// ===================== 11. 构建前端响应 =====================
	var orderItemsResp []*order.OrderItem
	for _, s := range itemSnapshots {
		orderItemsResp = append(orderItemsResp, &order.OrderItem{
			SkuId:       s.SkuId,
			SpuId:       s.SpuId,
			ProductName: s.SpuName,
			SkuName:     s.SkuName,
			Pic:         s.SkuPic,
			Price:       s.Price,
			Quantity:    s.Count,
			TotalAmount: s.TotalAmount,
		})
	}

	// ===================== 12. 生成结算token =====================
	settlementToken, err := l.svcCtx.IDGenerator.GenWithPrefix(constants.PREFIX_SETTLEMENT_TOKEN)
	if err != nil {
		l.Logger.Errorf("生成结算Token失败，error=%v", err)
		return nil, err
	}

	// ===================== 13. 缓存完整快照 =====================
	settlementData := SettlementData{
		UserId:          userId,
		AddressSnapshot: string(addrSnapshotStr),
		CouponId:        in.CouponId,
		ItemSnapshots:   itemSnapshots,
		TotalAmount:     totalAmount,
		FreightAmount:   freightAmount,
		CouponDiscount:  couponDiscount,
		PayAmount:       payAmount,
	}
	cacheJson, err := json.Marshal(settlementData)
	if err != nil {
		l.Logger.Errorf("序列化结算数据失败，error=%v", err)
		return nil, err
	}

	cacheKey := constants.SETTLEMENT_TOKEN_KEY + settlementToken
	ok, err := l.svcCtx.Client.SetNX(l.ctx, cacheKey, cacheJson, constants.SETTLEMENT_TOKEN_EXPIRE).Result()
	if err != nil || !ok {
		l.Logger.Errorf("缓存结算令牌失败，error=%v", err)
		return nil, err
	}

	l.Logger.Info("预下单成功")

	// ===================== 14. 返回结果 =====================
	return &order.PreOrderResponse{
		SettlementToken: settlementToken,
		Items:           orderItemsResp,
		Address:         addrSnapshot,
		TotalAmount:     totalAmount,
		FreightAmount:   freightAmount,
		CouponAmount:    couponDiscount,
		PayAmount:       payAmount,
	}, nil
}
