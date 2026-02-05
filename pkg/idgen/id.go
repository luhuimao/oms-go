package idgen

import "sync/atomic"

type Generator struct {
	id int64
}

func New() *Generator {
	return &Generator{id: 1000}
}

func (g *Generator) Next() int64 {
	return atomic.AddInt64(&g.id, 1)
}
