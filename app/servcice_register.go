package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/daemtri/begonia/bootstrap"
	"github.com/daemtri/begonia/di/box"
	"github.com/daemtri/begonia/di/box/flagvar"
	"github.com/daemtri/begonia/di/container"
	"github.com/daemtri/begonia/pkg/execx"
	"github.com/daemtri/begonia/pkg/netx"
	"github.com/daemtri/begonia/runtime"
	"github.com/daemtri/begonia/runtime/component"
)

func init() {
	execx.Register("discovery-sidecar", serviceRegister)
	if execx.Init() {
		os.Exit(0)
	}
}

func serviceRegister() {
	box.Provide[component.Discovery](&runtime.Builder[component.Discovery]{Name: "file"}, box.WithFlags("discovery"))

	var serviceEntryJSON string
	var ppid int
	var namespace string
	box.FlagSet().StringVar(&serviceEntryJSON, "service-entry", "", "")
	box.FlagSet().IntVar(&ppid, "ppid", 0, "")
	box.FlagSet().StringVar(&namespace, "namespace", "", "")

	box.UseInit(func(_ context.Context) error {
		runtime.SetNamespace(namespace)
		return nil
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	discovery, err := box.Build[component.Discovery](ctx)
	if err != nil {
		logger.Error("获取服务发现组件出错", "error", err)
		return
	}
	if ppid == 0 {
		logger.Error("必须指定父进程ID")
		return
	}

	var serviceEntry component.ServiceEntry
	if err := json.Unmarshal([]byte(serviceEntryJSON), &serviceEntry); err != nil {
		logger.Error("解析ServiceEntry出错", "error", err)
		return
	}
	if err := discovery.Register(ctx, serviceEntry); err != nil {
		logger.Error("服务注册出错", "error", err)
		return
	}
	for {
		// linux 已经注册了 PdeathSig,windows和mac不支持,暂时这样处理
		if os.Getppid() != ppid {
			// TODO: call DeRegister
			return
		}
		time.Sleep(time.Second)
	}
}

func initRegisterApp(ctx context.Context, discovery component.Discovery, servers []bootstrap.Server) error {
	for i := range servers {
		if !servers[i].Enabled() {
			continue
		}
		addr := servers[i].BroadCastAddr()
		parsedAddr, err := url.Parse(addr)
		if err != nil {
			return err
		}
		if parsedAddr.Hostname() == "" || parsedAddr.Hostname() == "0.0.0.0" {
			localIP, err := getBroadCastHost()
			if err != nil {
				return err
			}
			addr = fmt.Sprintf("%s://%s:%s", parsedAddr.Scheme, localIP, parsedAddr.Port())
		}
		runtime.AddServiceEndpoint(addr)
	}
	if len(runtime.GetServiceEndpoints()) == 0 {
		logger.Info("服务注册中断,未发现Endpoints")
		return nil
	}

	logger.Info("开始服务注册",
		"namespace", runtime.GetNamespace(),
		"name", runtime.GetServiceName(),
		"alias", runtime.GetServiceAlias(),
		"id", runtime.GetServiceID(),
		"endpoints", runtime.GetServiceEndpoints(),
		"metadata", runtime.GetServiceMetadata(),
	)

	if enableSideCarMode {
		ses, err := json.Marshal(runtime.GetServiceEntry())
		if err != nil {
			return err
		}

		box.FlagSet("xxx")
		args := []string{
			"discovery-sidecar",
			"-service-entry", string(ses),
			"-ppid", strconv.Itoa(os.Getpid()),
			"-discovery-name", getStringOpt("discovery-name"),
			"-namespace", runtime.GetNamespace(),
		}
		for _, opt := range getStringsOpt("discovery-opt") {
			args = append(args, "-discovery-opt", opt)
		}
		cmd := execx.Command(args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to run command %w", err)
		}
		go func() {
			if err := cmd.Wait(); err != nil {
				logger.Warn("failed to run command", "error", err)
			}
		}()
		return nil
	}
	return discovery.Register(ctx, runtime.GetServiceEntry())
}

func getBroadCastHost() (string, error) {
	if broadCastHost != "" {
		localIP := netx.ListLocalIP()
		if !slices.Contains(localIP, broadCastHost) {
			logger.Warn("设置的BroadCastHost地址不在本地局域网IP列表", "use", broadCastHost, "local", localIP)
		}
		return broadCastHost, nil
	}
	localIP, err := netx.FindLocalIP()
	if err != nil {
		return "", err
	}
	broadCastHost = localIP
	return broadCastHost, nil
}

func getStringOpt(name string) string {
	fs := flag.NewFlagSet("get-opt", flag.ContinueOnError)
	var val string
	fs.StringVar(&val, name, "", "")
	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}
	return val
}

func getStringsOpt(name string) []string {
	fs := flag.NewFlagSet("get-opt", flag.ContinueOnError)
	var val []string
	fs.Var(flagvar.Slice[string](&val), name, "")
	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}
	return val
}

type serviceRegisterRunable struct {
	servers   []bootstrap.Server
	discovery component.Discovery
}

func NewServiceRegisterRunable(ctx context.Context) *serviceRegisterRunable {
	return &serviceRegisterRunable{
		servers:   box.Invoke[container.Set[bootstrap.Server]](ctx),
		discovery: box.Invoke[component.Discovery](ctx),
	}
}

func (s *serviceRegisterRunable) Run(ctx context.Context) error {
	if err := initRegisterApp(ctx, s.discovery, s.servers); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}

func (s *serviceRegisterRunable) Enabled() bool {
	return true
}

func (s *serviceRegisterRunable) GracefulStop() {}
