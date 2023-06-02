package header

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cast"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func MetadataInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md.Copy())
	return handler(ctx, req)
}

func AddMetadata(ctx context.Context, MD metadata.MD) context.Context {
	for key, item := range MD {
		for _, val := range item {
			ctx = metadata.AppendToOutgoingContext(ctx, key, val)
		}
	}

	return ctx
}

func GetMetadata(ctx context.Context) (md metadata.MD, ok bool) {
	return metadata.FromOutgoingContext(ctx)
}

func getMetadataKey(ctx context.Context, key string) (result string, exist bool) {
	md, ok := GetMetadata(ctx)
	if !ok {
		return
	}

	data := md.Get(key)
	if len(data) > 0 {
		return data[0], true
	}

	return
}

func GetMetadataUID(ctx context.Context) (uid int64, exist bool) {
	result, ok := getMetadataKey(ctx, "userId")
	if !ok {
		return
	}

	u, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return
	}

	return u, true
}

func GetMetadataMediaId(ctx context.Context) (mediaId int32, exist bool) {
	result, ok := getMetadataKey(ctx, "mediaId")
	if !ok {
		return
	}

	mediaId = cast.ToInt32(result)
	return mediaId, true
}

func GetMetadataChannelId(ctx context.Context) (channelId int32, exist bool) {
	result, ok := getMetadataKey(ctx, "channelId")
	if !ok {
		return
	}

	channelId = cast.ToInt32(result)
	return channelId, true
}

func GetMetadataSubSN(ctx context.Context) (subSN string, exist bool) {
	result, ok := getMetadataKey(ctx, "subSN")
	if !ok {
		return
	}
	return result, true
}

func GetMetadataVersion(ctx context.Context) (version string, exist bool) {
	result, ok := getMetadataKey(ctx, "version")
	if !ok {
		return
	}

	return result, true
}

func GetMetadataUserAgent(ctx context.Context) (userAgent string, exist bool) {
	result, ok := getMetadataKey(ctx, "userAgent")
	if !ok {
		return
	}

	return result, true
}

func GetMetadataImei(ctx context.Context) (imei string, exist bool) {
	result, ok := getMetadataKey(ctx, "imei")
	if !ok {
		return
	}
	return result, true
}

func GetMetadataPhoneModel(ctx context.Context) (phoneModel string, exist bool) {
	result, ok := getMetadataKey(ctx, "phoneModel")
	if !ok {
		return
	}
	return result, true
}

func GetMetadataDistinctId(ctx context.Context) (result string, exist bool) {
	result, ok := getMetadataKey(ctx, "distinctId")
	if !ok {
		return
	}
	return result, true
}

func GetClusterId(ctx context.Context, appid uint8, userId int64) (clusterId uint8, exist bool) {
	result, ok := getMetadataKey(ctx, fmt.Sprintf("APP-%d-%d", appid, userId))
	if !ok {
		return
	}

	c, err := strconv.ParseUint(result, 10, 64)
	if err != nil {
		return
	}

	return uint8(c), true
}

func SetMetadataUID(ctx context.Context, uid int64) context.Context {
	v := strconv.FormatInt(uid, 10)
	return AddMetadata(ctx, metadata.Pairs("userId", v))
}

func SetMetadataParams(ctx context.Context, params map[string]string) context.Context {
	pairs := metadata.Pairs()
	for k, v := range params {
		pairs.Set(k, v)
	}
	return AddMetadata(ctx, pairs)
}

func SetClusterId(ctx context.Context, appid uint8, userId int64, clusterId uint8) context.Context {
	key := fmt.Sprintf("APP-%d-%d", appid, userId)
	v := strconv.Itoa(int(clusterId))
	return AddMetadata(ctx, metadata.Pairs(key, v))
}
