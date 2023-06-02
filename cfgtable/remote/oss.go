package remote

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssRemote struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	Bucket          string

	client *oss.Client
	bucket *oss.Bucket
}

func NewOssRemote(setting map[string]string) (Remote, error) {
	or := &OssRemote{}
	or.AccessKeyId = setting["AccessKeyId"]
	or.AccessKeySecret = setting["AccessKeySecret"]
	or.Endpoint = setting["Endpoint"]
	or.Bucket = setting["bucket"]
	var err error
	or.client, err = oss.New(or.Endpoint, or.AccessKeyId, or.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	or.bucket, err = or.client.Bucket(or.Bucket)
	if err != nil {
		return nil, err
	}
	return or, nil
}

func (o *OssRemote) Download(path string) (io.ReadCloser, error) {
	return o.bucket.GetObject(path)
}
