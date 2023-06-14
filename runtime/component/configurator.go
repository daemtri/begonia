package component

import (
	"context"
)

type configDecoder struct {
	raw    []byte
	decode func(raw []byte, x any) error
}

func NewConfigDecoder(raw []byte, decode func(raw []byte, x any) error) ConfigDecoder {
	return configDecoder{
		raw:    raw,
		decode: decode,
	}
}

func (cd configDecoder) Raw() []byte {
	return cd.raw
}

func (cd configDecoder) Decode(x any) error {
	return cd.decode(cd.raw, x)
}

type ConfigDecoder interface {
	Raw() []byte
	Decode(x any) error
}

type Configurator interface {
	Interface

	// ReadConfig 获取应用配置
	ReadConfig(ctx context.Context, name string) (ConfigDecoder, error)
	// WatchConfig 获取应用配置
	WatchConfig(ctx context.Context, name string) Iterator[ConfigDecoder]
}

type ConfigItem = struct {
	Key   string
	Value string
}
