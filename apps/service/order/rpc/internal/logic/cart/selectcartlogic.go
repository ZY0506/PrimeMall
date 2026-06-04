package cartlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type SelectCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSelectCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SelectCartLogic {
	return &SelectCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SelectCart 修改选中状态（单个商品）
func (l *SelectCartLogic) SelectCart(in *order.SelectCartRequest) (*order.Empty, error) {
	// 获取当前登录用户的id
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 查询购物车
	cart, err := l.svcCtx.CartModel.FindOneByUserIdSkuId(l.ctx, userId, in.SkuId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("购物车中没有此商品，skuId=%v", in.SkuId)
			return nil, nil
		}
		l.Logger.Errorf("查询购物车失败,error=%v", err)
		return nil, err
	}

	// 检查用户权限
	if cart.UserId != userId {
		l.Logger.Errorf("用户权限不足，用户ID=%v,购物车ID=%v", userId, cart.Id)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限修改购物车")
	}

	// 更新购物车状态
	cart.Selected = in.Selected
	err = l.svcCtx.CartModel.Update(l.ctx, cart)
	if err != nil {
		l.Logger.Errorf("更新购物车失败,error=%v", err)
		return nil, err
	}
	l.Logger.Info("更新购物车成功")

	return &order.Empty{}, nil
}
