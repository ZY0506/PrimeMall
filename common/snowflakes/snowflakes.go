package snowflakes

import (
	"context"
	"github.com/bwmarrin/snowflake"
)

type Generator struct {
	node *snowflake.Node
}

// NewGenerator 创建雪花ID生成器
func NewGenerator(nodeID int64) (*Generator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}
	return &Generator{
		node: node,
	}, nil
}

// NextID 生成雪花ID
func (s *Generator) NextID(ctx *context.Context) (id uint64, err error) {
	return uint64(s.node.Generate()), nil
}
