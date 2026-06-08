package productlogic

import (
	"context"
	"errors"
	"sort"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryInfoLogic {
	return &CategoryInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CategoryInfoLogic) CategoryInfo(in *product.IdReq) (*product.CategoryResp, error) {
	// 1. 查询当前分类
	curCategory, err := l.svcCtx.CategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("分类不存在, id=%d", in.Id)
			return nil, errorx.NewBizError(response.ErrCodeCategoryNotFound, "分类不存在")
		}
		l.Logger.Errorf("查询分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}
	if curCategory.DeleteAt.Valid {
		l.Logger.Errorf("分类已删除, id=%d", in.Id)
		return nil, errorx.NewBizError(response.ErrCodeCategoryNotFound, "分类不存在")
	} else if curCategory.Status != 1 {
		l.Logger.Errorf("分类已禁用, id=%d", in.Id)
		return nil, errorx.NewBizError(response.ErrCodeCategoryDisabled, "分类已禁用")
	}

	// 2. 查询所有子分类（递归查询，或一次性获取所有后代）
	children, err := l.svcCtx.CategoryModel.FindAllChildren(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("查询子分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	// 3. 构建当前分类的 proto 对象
	root := &product.Category{
		Id:       curCategory.Id,
		ParentId: curCategory.ParentId,
		Name:     curCategory.Name,
		Icon:     curCategory.Icon,
		Sort:     curCategory.Sort,
		Level:    curCategory.Level,
		Children: []*product.Category{},
	}

	// 4. 将子分类列表转换为 map，便于构建树
	childMap := make(map[uint64]*product.Category)
	for _, child := range *children {
		childMap[child.Id] = &product.Category{
			Id:       child.Id,
			ParentId: child.ParentId,
			Name:     child.Name,
			Icon:     child.Icon,
			Sort:     child.Sort,
			Level:    child.Level,
			Children: []*product.Category{},
		}
	}

	// 5. 将子分类挂载到对应的父节点下（从 root 开始，递归挂载）
	for _, node := range childMap {
		if node.ParentId == root.Id {
			root.Children = append(root.Children, node)
		} else {
			// 非直接子节点，找到其父节点挂载
			if parent, ok := childMap[node.ParentId]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return &product.CategoryResp{Category: root}, nil
}

// 递归排序函数
func sortChildren(cat *product.Category) {
	if len(cat.Children) == 0 {
		return
	}
	sort.Slice(cat.Children, func(i, j int) bool {
		if cat.Children[i].Sort == cat.Children[j].Sort {
			return cat.Children[i].Id < cat.Children[j].Id
		}
		return cat.Children[i].Sort < cat.Children[j].Sort
	})
	for _, child := range cat.Children {
		sortChildren(child)
	}
}
