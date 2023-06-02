package userredis

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/client"
)

var (
	Name string = "user"
)

func init() {
	box.Provide[*client.RedisClient](client.NewRedisClient, box.WithName(Name), box.WithFlags("redis-user"))
}
