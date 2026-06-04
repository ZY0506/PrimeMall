package productlogic

import (
	"context"
	"sort"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryTreeLogic {
	return &CategoryTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CategoryTree 分类
func (l *CategoryTreeLogic) CategoryTree(in *product.Empty) (*product.CategoryTreeResp, error) {
	// 获取所有分类
	list, err := l.svcCtx.CategoryModel.FindALL(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询分类出错，error=%v", err)
		return nil, err
	}

	// 构建分类map
	categoryMap := make(map[uint64]*product.Category)
	for _, v := range *list {
		categoryMap[v.Id] = &product.Category{
			Id:       v.Id,
			ParentId: v.ParentId,
			Name:     v.Name,
			Icon:     v.Icon,
			Sort:     v.Sort,
			Level:    v.Level,
			Children: []*product.Category{},
		}
	}

	// 构建分类树
	var categoryTree []*product.Category
	for _, category := range categoryMap {
		if category.ParentId == 0 {
			categoryTree = append(categoryTree, category)
		} else {
			// 不是父节点，则添加到父节点的子节点中
			// 找不到父节点，则忽略（视为整个分类都被禁用）
			if parent, ok := categoryMap[category.ParentId]; ok {
				parent.Children = append(parent.Children, category)
			}
		}
	}

	// 对每个节点的 children 按 sort 排序（可选，因为数据库已排序，但递归构建后可能乱序）
	var sortChildren func(node *product.Category)
	sortChildren = func(node *product.Category) {
		if len(node.Children) > 0 {
			sort.Slice(node.Children, func(i, j int) bool {
				if node.Children[i].Sort == node.Children[j].Sort {
					return node.Children[i].Id < node.Children[j].Id
				}
				return node.Children[i].Sort < node.Children[j].Sort
			})
			for _, child := range node.Children {
				sortChildren(child)
			}
		}
	}
	for _, root := range categoryTree {
		sortChildren(root)
	}
	l.Logger.Infof("获取分类树成功，分类树：%v", categoryTree)

	return &product.CategoryTreeResp{List: categoryTree}, nil
}
