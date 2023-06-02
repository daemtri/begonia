package main

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/cfgtable"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/config/apolloconfig"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/config/yamlconfig"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/bootstrap"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/client"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/config"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/provider/userredis"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/repository"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/server"
	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/service"
	"golang.org/x/exp/slog"
)

func main() {
	// register redis client
	box.Provide[contract.UserRepository](
		repository.NewUserRedisRepository,
		box.WithSelect[*client.RedisClient](userredis.Name),
	)
	// register service
	box.Provide[contract.Service](service.NewUserService, box.WithName("user"))
	box.Provide[contract.Service](service.NewConsumerService, box.WithName("consumer"))
	// register server
	box.Provide[contract.Server](server.NewHttpServer, box.WithName("http"), box.WithFlags("http"))
	// register logger
	box.Provide[*slog.Logger](slog.Default())
	// register app
	box.Provide[*bootstrap.App](bootstrap.NewApp)

	cfgtable.Init[*config.Aggregation]()

	// build and run
	if err := box.Bootstrap[*bootstrap.App](
		// The configuration has priority, the higher the priority of the closer
		yamlconfig.Init(),
		apolloconfig.Init(),
	); err != nil {
		panic(err)
	}
}
