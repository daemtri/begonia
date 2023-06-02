package types

import (
	"time"

	"golang.org/x/exp/constraints"
)

type BuildInType interface {
	constraints.Ordered | bool | time.Time
}

type Null[T BuildInType] struct {
	Value T
	Valid bool
}

func (n *Null[T]) Set(v T) {
	n.Value = v
	n.Valid = true
}
