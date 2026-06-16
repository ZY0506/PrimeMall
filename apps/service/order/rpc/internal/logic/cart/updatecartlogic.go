package cartlogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartLogic {
	return &UpdateCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCart 添加/更新商品（数量=0删除）
func (l *UpdateCartLogic) UpdateCart(in *order.UpdateCartRequest) (resp *order.Empty, err error) {
	var ok bool
	var userId uint64

	// 获取当前登录用户的id
	userId, err = ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 幂等校验
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.ORDER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等性校验失败,error=%v", err)
		return nil, err
	}
	if !ok {
		return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "请勿重复操作")
	}
	// TODO: 若后续失败，删除键（需确保删除操作的原子性，可配合 Lua 脚本）
	defer func() {
		if err != nil {
			l.svcCtx.Client.Del(l.ctx, idempotencyKey)
		}
	}()

	// 查询购物车
	cart, err := l.svcCtx.CartModel.FindOneByUserIdSkuId(l.ctx, userId, in.SkuId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			_, err = l.svcCtx.CartModel.Insert(l.ctx, &model.Cart{
				UserId:    userId,
				SkuId:     in.SkuId,
				Count:     uint64(in.Quantity),
				Selected:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				l.Logger.Errorf("添加商品失败,error=%v", err)
				return nil, err
			}
			l.Logger.Info("添加购物车成功")
			return &order.Empty{}, nil
		} else {
			l.Logger.Errorf("查询商品失败,error=%v", err)
			return nil, err
		}
	}

	// 检查用户权限
	if cart.UserId != userId {
		l.Logger.Errorf("用户权限不足，用户ID=%v,购物车ID=%v", userId, cart.Id)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限修改购物车")
	}

	// 更新购物车
	if in.Quantity == 0 {
		// 删除商品
		err = l.svcCtx.CartModel.Delete(l.ctx, cart.Id)
		if err != nil {
			l.Logger.Errorf("删除商品失败,error=%v", err)
			return nil, err
		}
	} else {
		// 更新商品
		cart.Count = uint64(in.Quantity)
		err = l.svcCtx.CartModel.Update(l.ctx, cart)
		if err != nil {
			l.Logger.Errorf("更新商品失败,error=%v", err)
			return nil, err
		}
	}
	l.Logger.Info("更新购物车成功")

	return &order.Empty{}, nil
}
