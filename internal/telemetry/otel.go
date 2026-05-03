package telemetry

import (
	"context"
	"net/http"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type OTel struct {
	meterProvider *sdkmetric.MeterProvider
	registry      *promclient.Registry
}

func New() (*OTel, error) {
	registry := promclient.NewRegistry()

	exporter, err := otelprom.New(
		otelprom.WithRegisterer(registry),
		otelprom.WithNamespace(""),
		otelprom.WithProducer(otelruntime.NewProducer()),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", "todo-api"),
			attribute.String("service.version", "0.1.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	if err := otelruntime.Start(
		otelruntime.WithMeterProvider(meterProvider),
	); err != nil {
		return nil, err
	}

	return &OTel{
		meterProvider: meterProvider,
		registry:      registry,
	}, nil
}

func (o *OTel) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(o.registry, promhttp.HandlerOpts{})
}

func (o *OTel) Shutdown(ctx context.Context) error {
	return o.meterProvider.Shutdown(ctx)
}
