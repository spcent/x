package httpclient

import (
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	traceName = "github.com/spcent/x/net/http"
)

var (
	tracer trace.Tracer
)

// nolint
func init() {
	tracer = otel.GetTracerProvider().Tracer(traceName, trace.WithInstrumentationVersion(contrib.Version()))
}
