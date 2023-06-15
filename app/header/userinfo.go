package header

import (
	"context"
	"fmt"
	"strconv"

	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"google.golang.org/grpc/metadata"
)

type UserInfo struct {
	md metadata.MD
}

func GetUserInfoFromIncomingCtx(ctx context.Context) *UserInfo {
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		panic(fmt.Errorf("no metadata in context"))
	}
	return &UserInfo{md: md}
}

func (u *UserInfo) get(key string) string {
	ret := u.md.Get("tenant_id")
	if len(ret) == 0 {
		panic(fmt.Errorf("no %s in metadata", key))
	}
	return ret[0]
}

func (u *UserInfo) GetTenantID() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("tenant_id"))))
}

func (u *UserInfo) GetUserID() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("user_id"))))
}

func (u *UserInfo) GetGameID() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("game_id"))))
}

func (u *UserInfo) GetSource() string {
	return u.get("source")
}

func (u *UserInfo) GetVersion() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("version"))))
}
