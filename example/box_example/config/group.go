package config

import (
	"github.com/daemtri/begonia/cfgtable"
	"github.com/daemtri/begonia/example/box_example/config/table"
)

type Aggregation struct {
	Level cfgtable.Config[table.Level] `flag:"level" usage:"等级配置"`
}

func (t *Aggregation) OnLoad(ctx cfgtable.Context) {
}
