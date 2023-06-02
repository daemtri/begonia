package remote

import (
	"fmt"
	"io"
)

var (
	registry = map[string]MakeRemote{
		"http": NewHttpRemote,
		"oss":  NewOssRemote,
	}
	defaultRemote Remote = Must(NewHttpRemote(map[string]string{}))
)

type Remote interface {
	Download(path string) (io.ReadCloser, error)
}

type MakeRemote func(settings map[string]string) (Remote, error)

func NewRemote(provider string, setting map[string]string) (Remote, error) {
	mr, ok := registry[provider]
	if !ok {
		return nil, fmt.Errorf("remote provider %s not exists", provider)
	}
	return mr(setting)
}

func SetDefault(provider string, setting map[string]string) error {
	r, err := NewRemote(provider, setting)
	if err != nil {
		return err
	}
	defaultRemote = r
	return nil
}

func Download(path string) (io.ReadCloser, error) {
	return defaultRemote.Download(path)
}

func Must(r Remote, e error) Remote {
	if e != nil {
		panic(e)
	}
	return r
}
