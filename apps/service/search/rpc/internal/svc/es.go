package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const productIndexName = "prime_mall_products"

// ProductDoc 商品ES文档结构
type ProductDoc struct {
	ID           uint64   `json:"id"`
	Name         string   `json:"name"`
	Brand        string   `json:"brand"`
	Description  string   `json:"description"`
	Cover        string   `json:"cover"`
	Price        int64    `json:"price"`
	Sales        int64    `json:"sales"`
	CategoryID   int64    `json:"category_id"`
	CategoryName string   `json:"category_name"`
	Suggest      []string `json:"suggest"`
}

// ESClient 封装ES操作
type ESClient struct {
	client *elasticsearch.Client
}

func NewESClient(addresses []string, username, password string) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch new client error: %w", err)
	}
	_, err = client.Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch info error: %w", err)
	}
	return &ESClient{client: client}, nil
}

// EnsureIndex 确保商品索引存在
func (es *ESClient) EnsureIndex(ctx context.Context) error {
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{productIndexName},
	}
	existsRes, err := existsReq.Do(ctx, es.client)
	if err != nil {
		return err
	}
	defer existsRes.Body.Close()
	if existsRes.StatusCode == 200 {
		return nil
	}

	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"ik_smart_analyzer": { "type": "custom", "tokenizer": "ik_smart" },
					"ik_max_word_analyzer": { "type": "custom", "tokenizer": "ik_max_word" }
				}
			}
		},
		"mappings": {
			"properties": {
				"id":             { "type": "long" },
				"name":           { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart", "fields": { "keyword": { "type": "keyword" } } },
				"brand":          { "type": "keyword" },
				"description":    { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
				"cover":          { "type": "keyword" },
				"price":          { "type": "long" },
				"sales":          { "type": "long" },
				"category_id":    { "type": "long" },
				"category_name":  { "type": "keyword" },
				"suggest":        { "type": "completion" }
			}
		}
	}`

	createReq := esapi.IndicesCreateRequest{
		Index: productIndexName,
		Body:  strings.NewReader(mapping),
	}
	createRes, err := createReq.Do(ctx, es.client)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		return fmt.Errorf("create index error: %s", createRes.String())
	}
	return nil
}

// IndexProduct 索引商品到ES
func (es *ESClient) IndexProduct(ctx context.Context, doc *ProductDoc) error {
	body, _ := json.Marshal(doc)
	req := esapi.IndexRequest{
		Index:      productIndexName,
		DocumentID: fmt.Sprintf("%d", doc.ID),
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}
	res, err := req.Do(ctx, es.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("index product error: %s", res.String())
	}
	return nil
}

// DeleteProduct 从ES删除商品
func (es *ESClient) DeleteProduct(ctx context.Context, productID uint64) error {
	req := esapi.DeleteRequest{
		Index:      productIndexName,
		DocumentID: fmt.Sprintf("%d", productID),
		Refresh:    "true",
	}
	res, err := req.Do(ctx, es.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete product error: %s", res.String())
	}
	return nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Total int64
	Hits  []ProductDoc
}

// SearchProducts 搜索商品
func (es *ESClient) SearchProducts(ctx context.Context, keyword string, categoryID int64, brand string,
	minPrice, maxPrice int64, sortBy, sortType string, page, size int) (*SearchResult, error) {

	mustQueries := make([]any, 0)
	if keyword != "" {
		mustQueries = append(mustQueries, map[string]any{
			"multi_match": map[string]any{
				"query":  keyword,
				"fields": []string{"name^3", "brand^2", "description"},
			},
		})
	}

	filterClauses := make([]any, 0)
	if categoryID > 0 {
		filterClauses = append(filterClauses, map[string]any{
			"term": map[string]any{"category_id": categoryID},
		})
	}
	if brand != "" {
		filterClauses = append(filterClauses, map[string]any{
			"term": map[string]any{"brand": brand},
		})
	}
	if minPrice > 0 {
		filterClauses = append(filterClauses, map[string]any{
			"range": map[string]any{"price": map[string]any{"gte": minPrice}},
		})
	}
	if maxPrice > 0 {
		filterClauses = append(filterClauses, map[string]any{
			"range": map[string]any{"price": map[string]any{"lte": maxPrice}},
		})
	}

	queryBody := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{},
		},
		"from": (page - 1) * size,
		"size": size,
	}

	boolQ := queryBody["query"].(map[string]any)["bool"].(map[string]any)
	if len(mustQueries) > 0 {
		boolQ["must"] = mustQueries
	}
	if len(filterClauses) > 0 {
		boolQ["filter"] = filterClauses
	}

	if sortBy != "" {
		order := "desc"
		if sortType == "asc" {
			order = "asc"
		}
		queryBody["sort"] = []map[string]any{
			{sortBy: map[string]string{"order": order}},
		}
	}

	body, _ := json.Marshal(queryBody)
	req := esapi.SearchRequest{
		Index: []string{productIndexName},
		Body:  strings.NewReader(string(body)),
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Score  float64    `json:"_score"`
				Source ProductDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := make([]ProductDoc, 0, len(result.Hits.Hits))
	for _, h := range result.Hits.Hits {
		doc := h.Source
		hits = append(hits, doc)
	}

	return &SearchResult{
		Total: result.Hits.Total.Value,
		Hits:  hits,
	}, nil
}

// Suggest 搜索建议
func (es *ESClient) Suggest(ctx context.Context, prefix string, size int) ([]string, error) {
	if size <= 0 {
		size = 5
	}

	queryBody := map[string]any{
		"suggest": map[string]any{
			"product_suggest": map[string]any{
				"prefix": prefix,
				"completion": map[string]any{
					"field": "suggest",
					"size":  size,
				},
			},
		},
	}

	body, _ := json.Marshal(queryBody)
	req := esapi.SearchRequest{
		Index: []string{productIndexName},
		Body:  strings.NewReader(string(body)),
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("suggest error: %s", res.String())
	}

	var result struct {
		Suggest map[string][]struct {
			Options []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"suggest"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	suggestions := make([]string, 0)
	if options, ok := result.Suggest["product_suggest"]; ok && len(options) > 0 {
		for _, opt := range options[0].Options {
			suggestions = append(suggestions, opt.Text)
		}
	}
	return suggestions, nil
}
