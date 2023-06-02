package tracing

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var log = logx.GetLogger("grpcx/tracing")

type Factory struct {
	Endpoint string `flag:"endpoint" default:"http://localhost:14268/api/traces" usage:"jaeger 服务地址"`
	Exporter string `flag:"exporter" default:"stdout" usage:"支持 stdout|jaeger"`
	Sample   string `flag:"sample" default:"never" usage:"dynamic: 可选,always,never,dynamic"`

	ServiceNamespace  string
	ServiceName       string
	ServiceVersion    string
	ServiceInstanceID string
}

// NewTracerProvider 创建 TracerProvider
func (th *Factory) NewTracerProvider(attributes ...attribute.KeyValue) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	switch th.Exporter {
	case "jaeger":
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(th.Endpoint)))
	default:
		exporter, err = stdout.New(stdout.WithPrettyPrint())
	}
	if err != nil {
		return nil, err
	}

	// see: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md
	attributes = append(attributes, []attribute.KeyValue{
		semconv.ServiceNameKey.String(th.ServiceName),
		semconv.ServiceInstanceIDKey.String(th.ServiceInstanceID),
		semconv.ServiceNamespaceKey.String(th.ServiceNamespace),
		semconv.ServiceVersionKey.String(th.ServiceVersion),
	}...)

	var sample sdktrace.Sampler
	switch th.Sample {
	case "always":
		sample = sdktrace.AlwaysSample()
	case "never":
		sample = sdktrace.NeverSample()
	case "dynamic":
		sample = dynamicSample()
	default:
		sample = sdktrace.NeverSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sample),
		sdktrace.WithResource(
			resource.NewWithAttributes(semconv.SchemaURL, attributes...),
		),
		sdktrace.WithBatcher(exporter),
	)
	log.Debug("构建tracing", "config", th)
	return tp, nil
}
