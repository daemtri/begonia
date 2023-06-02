package remote

import (
	"fmt"
	"io"
	"net/http"
)

type HttpRemote struct {
	basePath string
}

func NewHttpRemote(setting map[string]string) (Remote, error) {
	return &HttpRemote{basePath: setting["base_path"]}, nil
}

func (hr *HttpRemote) Download(path string) (io.ReadCloser, error) {
	resp, err := http.Get(hr.basePath + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http.Get status=%s", resp.Status)
	}
	return resp.Body, nil
}
