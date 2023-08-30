package cfgtable

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/daemtri/begonia/cfgtable/remote"
)

var _ ConfigInterface = &Config[string]{}

type ConfigInterface interface {
	URI() string
	Fetch() ([]byte, error)
	Parse([]byte) error
}

type JSON[T any] struct {
	Version string `json:"Version"`
	Items   []T    `json:"Items"`
}

func (j *JSON[T]) Reset() {
	j.Version = ""
	j.Items = []T{}
}

type Config[T any] struct {
	JSON[T]

	// URL 配置远程下载地址
	URL string `flag:"url" usage:"远端下载路径"`
}

func (c *Config[T]) URI() string {
	return c.URL
}

func (c *Config[T]) secret() string {
	return defaultSecret
}

func (c *Config[T]) Parse(data []byte) error {
	mux.Lock()
	defer mux.Unlock()
	if err := json.Unmarshal(data, &c.JSON); err != nil {
		return fmt.Errorf("secret:%s, json解析失败: %w", c.secret(), err)
	}
	return nil
}

func (c *Config[T]) Fetch() ([]byte, error) {
	if c.URL == "" {
		return nil, errors.New("远程加载配置失败,URL为空")
	}
	rd, err := remote.Download(c.URL)
	if err != nil {
		return nil, err
	}
	defer rd.Close()

	data, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if secret := c.secret(); secret != "" {
		data, err = AesCtrDecrypt(data, []byte(secret))
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
