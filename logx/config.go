package logx

type Config struct {
	// 默认日志output
	Default string `flag:"default"  usage:"default handler"`

	// Output 日志输出，格式为 NAME=DSN
	// 示例：
	//  log1=./log/some-app-name.log?level=info&rotate=true&max-size=50&local-time=true&max-backups=10&max-age=30&compress=true
	//  log2=stdout?level=debug&fromat=console&addsource=true
	Handler map[string]string `flag:"handler" usage:"Log-driven configuration"`
	// Logger 日志驱动配置，用于为某些日志指定 handler
	// 如：log := logx.GetLogger("main"), 则指定其output格式如下
	//	main=log1+log2
	Logger map[string]string `flag:"logger" default:"" usage:"Component log configuration"`
}

var (
	handlerManagers map[string]*handlerManager = make(map[string]*handlerManager)
)

func InitConfig(cfg Config) error {
	for k, v := range cfg.Handler {
		hm, ok := handlerManagers[k]
		if ok {
			err := hm.configure(v)
			if err != nil {
				return err
			}
			continue
		}
		hm, err := newHandlerManager(v)
		if err != nil {
			return err
		}
		handlerManagers[k] = hm
	}
	return nil
}
