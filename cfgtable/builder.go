package cfgtable

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"git.bianfeng.com/stars/wegame/wan/wanx/cfgtable/remote"
)

type RemoteBuilder[T Aggregation] struct {
	Builder[T]     `flag:""`
	RemoteProvider string            `flag:"remote_provider" default:"http" usage:"远程下载提供者"`
	RemoteSetting  map[string]string `flag:"remote_setting" usage:"远程下载提供者配置"`
	Secret         string            `flag:"secret" usage:"配置文件默认解密密钥,如果指定了该配置,且具体配置文件未指定scret,则默认使用该配置"`
}

func NewRemoteBuilder[T Aggregation]() (*RemoteBuilder[T], error) {
	b, err := NewBuilder[T]()
	if err != nil {
		return nil, err
	}
	return &RemoteBuilder[T]{
		Builder: *b,
	}, nil
}

func (rb *RemoteBuilder[T]) Retrofit() error {
	defaultSecret = rb.Secret
	if err := remote.SetDefault(rb.RemoteProvider, rb.RemoteSetting); err != nil {
		return err
	}
	return rb.Builder.Retrofit()
}

func (rb *RemoteBuilder[T]) Build(_ context.Context) (T, error) {
	return rb.CC, rb.Retrofit()
}

type Builder[T Aggregation] struct {
	CC T `flag:"" usage:"配置表"`

	refTyp reflect.Type
	refVal reflect.Value
	// lastURL存储了上一次的URL，以变对比是否发生过变动
	lastURIs map[string]string
}

func NewBuilder[T Aggregation]() (*Builder[T], error) {
	b := &Builder[T]{
		lastURIs: map[string]string{},
	}
	refTyp := reflect.TypeOf(b.CC)
	if refTyp.Kind() != reflect.Pointer && refTyp.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("configTable初始化ConfigCollection类型必须是结构体指针,收到%s", refTyp)
	}
	b.refTyp = refTyp.Elem()
	b.refVal = reflect.New(b.refTyp).Elem()
	b.CC, _ = b.refVal.Addr().Interface().(T)
	return b, nil
}

func (b *Builder[T]) Retrofit() error {
	nf := b.refTyp.NumField()
	changes := map[ConfigInterface]bool{}
	for i := 0; i < nf; i++ {
		if !b.refTyp.Field(i).IsExported() {
			continue
		}

		var cfg ConfigInterface
		var ok bool
		field := b.refVal.Field(i)
		if field.Kind() == reflect.Struct {
			cfg, ok = field.Addr().Interface().(ConfigInterface)
		} else if field.Kind() == reflect.Pointer && b.refTyp.Field(i).Type.Elem().Kind() == reflect.Struct {
			if field.IsZero() {
				field.Set(reflect.New(b.refTyp.Field(i).Type.Elem()))
			}
			cfg, ok = field.Interface().(ConfigInterface)
		}
		if !ok {
			continue
		}
		name := strings.TrimSuffix(strings.TrimSuffix(b.refTyp.Field(i).Tag.Get("flag"), "nested"), ",")
		if cfg.URI() == b.lastURIs[name] {
			continue
		}
		if b.lastURIs[name] != "" {
			b.loadConfigFromUri(name, cfg)
		} else {
			b.initConfig(name, cfg)
		}
		b.lastURIs[name] = cfg.URI()
		changes[cfg] = true
	}
	b.CC.OnLoad(Context{
		changes: changes,
	})
	return nil
}

func (b *Builder[T]) loadConfigFromCache(name string, cfg ConfigInterface) error {
	data, err := loadFromBackupFile(name)
	if err != nil {
		return err
	}
	if err := cfg.Parse(data); err != nil {
		return err
	}
	return nil
}

func (b *Builder[T]) initConfig(name string, cfg ConfigInterface) {
	// 第一次会先尝试URL，再尝试缓存
	data, err := cfg.Fetch()
	if err == nil {
		err = cfg.Parse(data)
	}
	if err == nil {
		if err := saveToBackupFile(name, data); err != nil {
			log.Error("cfg sav backup failed",
				"name", name,
				"uri1", cfg.URI(),
			)
			os.Exit(1)
		}
		return
	}
	if err2 := b.loadConfigFromCache(name, cfg); err2 != nil {
		// 从缓存中读取
		log.Error("cfg fetach or parse failed from uri and backup file",
			"name", name,
			"uri2", cfg.URI(),
			"uri_error", err,
			"backup_error", err2,
		)
		os.Exit(1)
	}
	// 从缓存中读取
	log.Warn("cfg fetach or parse failed from uri, succeed from backup file",
		"name", name,
		"uri3", cfg.URI(),
		"uri_error", err,
	)
}

func (b *Builder[T]) loadConfigFromUri(name string, cfg ConfigInterface) {
	// 第一次会先尝试URL，再尝试缓存
	data, err := cfg.Fetch()
	if err == nil {
		err = cfg.Parse(data)
	}
	if err == nil {
		if err := saveToBackupFile(name, data); err != nil {
			log.Warn("cfg sav backup failed",
				"name", name,
				"uri", cfg.URI(),
			)
		}
		return
	}
}

func (b *Builder[T]) Build(_ context.Context) (T, error) {
	return b.CC, b.Retrofit()
}
