package logx

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/exp/slog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	handlerAgents  = make(map[string]*handlerAgent)
	loggerHandlers = make(map[string]slog.Handler)
)

func createHandlerAgent(name string) (*handlerAgent, error) {
	handlerStr := globalConfig.Handler[name]
	if handlerStr == "" {
		return nil, nil
	}
	var agentOpt handlerAgentOptions
	if err := agentOpt.parse(handlerStr); err != nil {
		return nil, err
	}
	agent, err := newHandlerAgent(&agentOpt)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func createLogHandler(logger string, getHandlerAgent func(name string) *handlerAgent) (slog.Handler, error) {
	loggerHandlerStr := globalConfig.Logger[logger]
	if loggerHandlerStr == "" {
		return slog.Default().Handler(), nil
	}
	handlerSlices := strings.Split(loggerHandlerStr, "+")
	if len(handlerSlices) == 1 {
		handlerName := strings.TrimSpace(handlerSlices[0])
		agent := getHandlerAgent(handlerName)
		if agent == nil {
			return slog.Default().Handler(), nil
		}
		return agent, nil
	}
	handlers := make([]slog.Handler, 0, len(handlerSlices))
	for i := range handlerSlices {
		handlerName := strings.TrimSpace(handlerSlices[i])
		agent := getHandlerAgent(handlerName)
		if agent == nil {
			continue
		}
		handlers = append(handlers, agent)
	}
	return handlerGroup(handlers), nil
}

func getLoggerHandler(logger string) slog.Handler {
	handler, ok := loggerHandlers[logger]
	if !ok {
		handler, err := createLogHandler(logger, func(name string) *handlerAgent {
			handler, ok := handlerAgents[name]
			if !ok {
				agent, err := createHandlerAgent(name)
				if err != nil {
					panic(err)
				}
				handlerAgents[name] = agent
				return agent
			}
			return handler
		})
		if err != nil {
			panic(err)
		}
		loggerHandlers[logger] = handler
		return handler
	}
	return handler
}

type handlerAgentOptions struct {
	path       string
	addSource  bool
	level      string
	format     string
	maxsize    int
	maxage     int
	maxbackups int
}

func (h *handlerAgentOptions) parse(u string) error {
	var err error
	cfgUrl, err := url.Parse(u)
	if err != nil {
		return err
	}
	h.path = cfgUrl.Path
	h.addSource, err = getBoolFromUrl(cfgUrl, "addsource")
	if err != nil {
		return err
	}
	h.level = cfgUrl.Query().Get("level")
	h.format = strings.ToUpper(cfgUrl.Query().Get("format"))
	h.maxsize, err = getIntFromUrl(cfgUrl, "maxsize", 100)
	if err != nil {
		return err
	}
	h.maxage, err = getIntFromUrl(cfgUrl, "maxage", 7)
	if err != nil {
		return err
	}
	h.maxbackups, err = getIntFromUrl(cfgUrl, "maxbackups", 3)
	if err != nil {
		return err
	}
	return nil
}

type handlerAgent struct {
	dynamicAddSourcerHandler

	agentOpt *handlerAgentOptions
	opt      slog.HandlerOptions
	leveler  dynamicLeveler
	writer   dynamicWriter
}

func newHandlerAgent(opt *handlerAgentOptions) (*handlerAgent, error) {
	ha := &handlerAgent{}
	return ha, ha.configure(opt)
}

func (ha *handlerAgent) configure(opt *handlerAgentOptions) error {
	if reflect.DeepEqual(ha.agentOpt, opt) {
		return nil
	}
	if err := ha.leveler.setLevel(opt.level); err != nil {
		return err
	}
	if err := ha.writer.configure(opt.path, opt.maxsize, opt.maxbackups, opt.maxage, true, false); err != nil {
		return err
	}
	ha.sourcer.setAddSource(opt.addSource)
	if ha.opt.Level == nil {
		ha.opt.Level = &ha.leveler
	}
	ha.opt.AddSource = false
	if ha.Handler == nil || opt.format != ha.agentOpt.format {
		if opt.format == "JSON" {
			ha.Handler = slog.NewJSONHandler(&ha.writer, &ha.opt)
		} else if opt.format == "TEXT" {
			ha.Handler = slog.NewTextHandler(&ha.writer, &ha.opt)
		} else {
			return errors.New("invalid format")
		}
	}
	ha.agentOpt = opt
	return nil
}

type dynamicAddSourcerHandler struct {
	slog.Handler
	sourcer dynamicAddSourcer
}

func (h *dynamicAddSourcerHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.sourcer.addSource() {
		r.AddAttrs(slog.String(slog.SourceKey, "xxx"))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *dynamicAddSourcerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &dynamicAddSourcerHandler{
		sourcer: h.sourcer,
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h *dynamicAddSourcerHandler) WithGroup(name string) slog.Handler {
	return &dynamicAddSourcerHandler{
		sourcer: h.sourcer,
		Handler: h.Handler.WithGroup(name),
	}
}

type dynamicWriter struct {
	writer *lumberjack.Logger
}

func (w *dynamicWriter) configure(filename string, maxSize, maxBackups, maxAge int, localTime, compress bool) error {
	if w.writer == nil {
		w.writer = &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge, //days
			LocalTime:  localTime,
			Compress:   compress,
		}
		return nil
	}
	if filename == w.writer.Filename &&
		maxSize == w.writer.MaxSize &&
		maxBackups == w.writer.MaxBackups &&
		maxAge == w.writer.MaxAge &&
		localTime == w.writer.LocalTime &&
		compress == w.writer.Compress {
		return nil
	}

	w.writer.Filename = filename
	w.writer.MaxSize = maxSize
	w.writer.MaxBackups = maxBackups
	w.writer.MaxAge = maxAge
	w.writer.LocalTime = localTime
	w.writer.Compress = compress
	return w.writer.Rotate()
}

func (w *dynamicWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

type dynamicLeveler struct {
	level slog.Level
}

func (l *dynamicLeveler) setLevel(name string) error {
	switch strings.ToUpper(name) {
	case "DEBUG":
		l.level = slog.LevelDebug
	case "INFO":
		l.level = slog.LevelInfo
	case "WARN":
		l.level = slog.LevelWarn
	case "ERROR":
		l.level = slog.LevelError
	default:
		return errors.New("unknown name")
	}
	return nil
}

func (l *dynamicLeveler) Level() slog.Level {
	return l.level
}

type dynamicAddSourcer struct {
	addSourceValue bool
}

func newDynamicAddSourcer(v bool) *dynamicAddSourcer {
	return &dynamicAddSourcer{addSourceValue: v}
}

func (s *dynamicAddSourcer) setAddSource(addsource bool) {
	s.addSourceValue = addsource
}

func (s *dynamicAddSourcer) addSource() bool {
	return s.addSourceValue
}

func (s *dynamicAddSourcer) frame(r slog.Record) runtime.Frame {
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	return f
}

func getBoolFromUrl(u *url.URL, name string) (bool, error) {
	s := u.Query().Get(name)
	if s == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return v, nil
}

func getIntFromUrl(u *url.URL, name string, def int) (int, error) {
	str := u.Query().Get(name)
	if str == "" {
		return def, nil
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

type handlerGroup []slog.Handler

func (hg handlerGroup) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range hg {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (hg handlerGroup) Handle(ctx context.Context, record slog.Record) error {
	var err error
	for _, h := range hg {
		if err2 := h.Handle(ctx, record); err2 != nil {
			err = errors.Join(err, err2)
		}
	}
	return err
}

func (hg handlerGroup) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(hg))
	for i, h := range hg {
		handlers[i] = h.WithAttrs(attrs)
	}
	return handlerGroup(handlers)
}

func (hg handlerGroup) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(hg))
	for i, h := range hg {
		handlers[i] = h.WithGroup(name)
	}
	return handlerGroup(handlers)
}
