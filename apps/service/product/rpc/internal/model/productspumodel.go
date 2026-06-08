package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"time"
)

var _ ProductSpuModel = (*customProductSpuModel)(nil)

type (
	// ProductSpuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductSpuModel.
	ProductSpuModel interface {
		productSpuModel
		customProductSpu
		withSession(session sqlx.Session) ProductSpuModel
	}

	customProductSpuModel struct {
		*defaultProductSpuModel
	}
	customProductSpu interface {
		FindListByFilter(ctx context.Context, filters *ListFilters) ([]*ProductItem, int64, error)
		InsertTx(ctx context.Context, session sqlx.Session, data *ProductSpu) (sql.Result, error)
		UpdateTx(ctx context.Context, session sqlx.Session, data *ProductSpu) error
		DeleteTx(ctx context.Context, session sqlx.Session, id uint64) error
	}

	// AttrFilter 定义属性过滤项
	AttrFilter struct {
		Name   string
		Values []string
	}

	// ListFilters 定义过滤参数结构体
	ListFilters struct {
		CategoryId uint64
		Brand      string
		MinPrice   int64 // 单位：分
		MaxPrice   int64
		Keyword    string
		Attrs      []AttrFilter // 解析后的属性过滤
		IsNew      bool
		SortBy     string // price, sales, created_at
		SortType   string // asc, desc
		Page       int64
		PageSize   int64
	}
	ProductItem struct {
		Id        uint64 `db:"id"`
		Name      string `db:"name"`
		Brand     string `db:"brand"`
		Desc      string `db:"desc"`
		Cover     string `db:"cover"`
		Price     int64  `db:"price"` // 单位：分
		Sales     int64  `db:"sales"`
		ShowSales int64  `db:"show_sales"`
	}
)

// NewProductSpuModel returns a model for the database table.
func NewProductSpuModel(conn sqlx.SqlConn) ProductSpuModel {
	return &customProductSpuModel{
		defaultProductSpuModel: newProductSpuModel(conn),
	}
}

func (m *customProductSpuModel) withSession(session sqlx.Session) ProductSpuModel {
	return NewProductSpuModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customProductSpuModel) InsertTx(ctx context.Context, session sqlx.Session, data *ProductSpu) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, productSpuRowsExpectAutoSet)
	ret, err := session.ExecCtx(ctx, query, data.CategoryId, data.Name, data.Brand, data.Desc, data.Content, data.MainPic, data.SubPics, data.VideoUrl, data.Status, data.SalesCount, data.VirtualSales, data.FreightTemplateId, data.DeleteAt)
	return ret, err
}

func (m *customProductSpuModel) UpdateTx(ctx context.Context, session sqlx.Session, data *ProductSpu) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, productSpuRowsWithPlaceHolder)
	_, err := session.ExecCtx(ctx, query, data.CategoryId, data.Name, data.Brand, data.Desc, data.Content, data.MainPic, data.SubPics, data.VideoUrl, data.Status, data.SalesCount, data.VirtualSales, data.FreightTemplateId, data.DeleteAt, data.Id)
	return err
}

func (m *customProductSpuModel) DeleteTx(ctx context.Context, session sqlx.Session, id uint64) error {
	query := fmt.Sprintf("update %s set delete_at = ? where `id` = ?", m.table)
	_, err := session.ExecCtx(ctx, query, time.Now(), id)
	return err
}

func (m *defaultProductSpuModel) FindListByFilter(ctx context.Context, filters *ListFilters) ([]*ProductItem, int64, error) {
	// 1. 构建过滤条件
	whereClauses := []string{"p.status = 1", "p.delete_at IS NULL"} // 默认只查上架商品
	var args []interface{}

	if filters.CategoryId > 0 {
		whereClauses = append(whereClauses, "p.category_id = ?")
		args = append(args, filters.CategoryId)
	}
	if filters.Brand != "" {
		whereClauses = append(whereClauses, "p.brand = ?")
		args = append(args, filters.Brand)
	}
	if filters.Keyword != "" {
		whereClauses = append(whereClauses, "p.name LIKE ?")
		args = append(args, "%"+filters.Keyword+"%")
	}
	if filters.IsNew {
		whereClauses = append(whereClauses, "p.created_at > DATE_SUB(NOW(), INTERVAL 7 DAY)")
	}
	// 价格过滤
	if filters.MinPrice > 0 || filters.MaxPrice > 0 {
		priceSubQuery := "EXISTS (SELECT 1 FROM product_sku s WHERE s.spu_id = p.id AND s.status = 1"
		if filters.MinPrice > 0 {
			priceSubQuery += " AND s.price >= ?"
			args = append(args, float64(filters.MinPrice)/100.0)
		}
		if filters.MaxPrice > 0 {
			priceSubQuery += " AND s.price <= ?"
			args = append(args, float64(filters.MaxPrice)/100.0)
		}
		priceSubQuery += ")"
		whereClauses = append(whereClauses, priceSubQuery)
	}
	//// 多属性交集过滤
	//if len(filters.Attrs) > 0 {
	//	attrSubQuery := "EXISTS (SELECT 1 FROM product_sku s2 WHERE s2.spu_id = p.id AND s2.status = 1"
	//	for key, value := range filters.Attrs {
	//		attrSubQuery += fmt.Sprintf(" AND JSON_EXTRACT(s2.spec_data, '$.%s') = ?", key)
	//		args = append(args, value)
	//	}
	//	attrSubQuery += ")"
	//	whereClauses = append(whereClauses, attrSubQuery)
	//}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 2. 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM product_spu p %s", whereSQL)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*ProductItem{}, 0, nil
	}

	// 3. 构建排序子句
	orderClause := "ORDER BY p.id DESC" // 默认按 id 倒序
	switch filters.SortBy {
	case "price":
		orderClause = fmt.Sprintf("ORDER BY min_price %s, p.id %s", filters.SortType, filters.SortType)
	case "sales":
		orderClause = fmt.Sprintf("ORDER BY p.sales_count %s, p.id %s", filters.SortType, filters.SortType)
	case "created_at":
		orderClause = fmt.Sprintf("ORDER BY p.created_at %s, p.id %s", filters.SortType, filters.SortType)
	}

	// 4. 分页参数
	offset := (filters.Page - 1) * filters.PageSize
	limit := filters.PageSize

	// 5. 主查询 SQL（价格排序需要特殊处理）
	var querySQL string
	sqlFields := `
        p.id,
        p.name,
        p.brand,
		p.desc,
        p.main_pic AS cover,
        CAST(COALESCE(MIN(s.price), 0) * 100 AS SIGNED) AS price,
        p.sales_count AS sales,
        (p.sales_count + p.virtual_sales) AS show_sales
    `

	// 按价格排序：必须在一个查询内完成 GROUP BY + 排序
	querySQL = fmt.Sprintf(`
		SELECT %s
		FROM product_spu p
		LEFT JOIN product_sku s ON p.id = s.spu_id AND s.status = 1
		%s
		GROUP BY p.id
		%s
		LIMIT ? OFFSET ?
	`, sqlFields, whereSQL, orderClause)

	// 6. 执行查询
	queryArgs := append(args, limit, offset)
	var items []*ProductItem
	if err := m.conn.QueryRowsCtx(ctx, &items, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
