package config

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/cfgtable"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/config/table"
)

type Aggregation struct {
	Level cfgtable.Config[table.Level] `flag:"level" usage:"等级配置"`
}

func (t *Aggregation) OnLoad(ctx cfgtable.Context) {
}
