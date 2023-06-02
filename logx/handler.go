package logx

import (
	"context"
	"errors"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/exp/slog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ slog.Handler = &handler{}

type handlerManager struct {
	addSource *dynamicAddSourcer
	level     *dynamicLeveler
	writer    *dynamicWriter
	handler   *handler
}

func newHandlerManager(u string) (*handlerManager, error) {
	hm := &handlerManager{}
	return hm, hm.configure(u)
}

// configure configures the handlerManager.
// url = "/path/to/file?level=debug&maxsize=100&maxage=7&maxbackups=3&format=json&addsource=true"
func (hm *handlerManager) configure(u string) error {
	cfgUrl, err := url.Parse(u)
	if err != nil {
		return err
	}
	v, err := getBoolFromUrl(cfgUrl, "addsource")
	if err != nil {
		return err
	}
	if hm.addSource == nil {
		hm.addSource = newDynamicAddSourcer(v)
	} else {
		hm.addSource.setAddSource(v)
	}

	level := cfgUrl.Query().Get("level")
	if hm.level == nil {
		hm.level = newDynamicLeveler()
	}
	if err := hm.level.setLevel(level); err != nil {
		return err
	}

	filename := cfgUrl.Host + cfgUrl.Path
	maxsize, err := getIntFromUrl(cfgUrl, "maxsize", 100)
	if err != nil {
		return err
	}
	maxage, err := getIntFromUrl(cfgUrl, "maxage", 7)
	if err != nil {
		return err
	}
	maxbackups, err := getIntFromUrl(cfgUrl, "maxbackups", 3)
	if err != nil {
		return err
	}
	localtime, err := getBoolFromUrl(cfgUrl, "localtime")
	if err != nil {
		return err
	}
	compress, err := getBoolFromUrl(cfgUrl, "compress")
	if err != nil {
		return err
	}
	if hm.writer == nil {
		hm.writer = newDynamicWriter(filename, maxsize, maxbackups, maxage, localtime, compress)
	} else {
		hm.writer.retrofit(filename, maxsize, maxbackups, maxage, localtime, compress)
	}
	if hm.handler == nil {
		format := strings.ToUpper(cfgUrl.Query().Get("format"))
		if format == "JSON" {
			hm.handler = &handler{
				addSource: hm.addSource,
				Handler:   slog.HandlerOptions{Level: hm.level}.NewJSONHandler(hm.writer),
			}
		} else if format == "TEXT" {
			hm.handler = &handler{
				addSource: hm.addSource,
				Handler:   slog.HandlerOptions{Level: hm.level}.NewTextHandler(hm.writer),
			}
		}
	}

	return nil
}

func (hm *handlerManager) getHandler() slog.Handler {
	return hm.handler
}

type handler struct {
	slog.Handler
	addSource *dynamicAddSourcer
}

func (h handler) Handle(ctx context.Context, r slog.Record) error {
	if h.addSource.addSource() {
		r.AddAttrs(slog.String(slog.SourceKey, "xxx"))
	}
	return h.Handler.Handle(ctx, r)
}

func (h handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler{
		addSource: h.addSource,
		Handler:   h.Handler.WithAttrs(attrs),
	}
}

func (h handler) WithGroup(name string) slog.Handler {
	return handler{
		addSource: h.addSource,
		Handler:   h.Handler.WithGroup(name),
	}
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

type dynamicWriter struct {
	writer *lumberjack.Logger
}

func newDynamicWriter(filename string, maxSize, maxBackups, maxAge int, localTime, compress bool) *dynamicWriter {
	dw := &dynamicWriter{}
	dw.writer = &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, //days
		LocalTime:  localTime,
		Compress:   compress,
	}

	return dw
}

func (w *dynamicWriter) retrofit(filename string, maxSize, maxBackups, maxAge int, localTime, compress bool) error {
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

func newDynamicLeveler() *dynamicLeveler {
	return &dynamicLeveler{
		level: slog.LevelInfo,
	}
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
