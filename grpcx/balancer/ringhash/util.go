package ringhash

import (
	"context"
	"github.com/cespare/xxhash/v2"
)

const (
	// MetadataHashKey 是grpc metadata key，
	// 对应的value形式为 KEY=VALUE, KEY和VALUE从SubConnInfo的Address的Metadata从获取
	// example:
	//		RingHash-Key: userid=123456
	MetadataHashKey = "RingHash-Key"
)

type RequestHashKey struct{}

var getRequestHash = func(ctx context.Context) uint64 {
	requestHashString, _ := ctx.Value(RequestHashKey{}).(string)
	return xxhash.Sum64String(requestHashString)
}

// GetRequestHashForTesting returns the request hash in the context; to be used
// for testing only.
func GetRequestHashForTesting(ctx context.Context) uint64 {
	return getRequestHash(ctx)
}

// SetRequestHash adds the request hash to the context for use in Ring Hash Load
// Balancing.
func SetRequestHash(ctx context.Context, requestHash string) context.Context {
	return context.WithValue(ctx, RequestHashKey{}, requestHash)
}

func SetRequestHashGetFunc(fn func(ctx context.Context) uint64) {
	getRequestHash = fn
}
