package otel

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ServiceName = semconv.ServiceNameKey.String("my-awesome-service")

func InitOtel(ctx context.Context, collectorAddr string) (Closers, error) {
	closers := Closers{}
	conn, err := initConn(collectorAddr)
	if err != nil {
		log.Fatal(err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(ServiceName))
	if err != nil {
		return closers, err
	}

	traceShutdownHook, err := initTracerProvider(ctx, res, conn)
	if err != nil {
		return closers, err
	}
	closers.TraceCloser = traceShutdownHook

	metricShutdownHook, err := initMeterProvider(ctx, res, conn)
	if err != nil {
		return closers, err
	}
	closers.MetricCloser = metricShutdownHook

	return closers, nil
}

func Tracer() trace.Tracer {
	return otel.Tracer("my-tracer")
}

func Meter() metric.Meter {
	return otel.Meter("my-metrics")
}

func initConn(collectorAddr string) (*grpc.ClientConn, error) {
	slog.Debug("connecting to otel collector", slog.String("addr", collectorAddr))
	conn, err := grpc.NewClient(
		collectorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish a connections to %s: %w", collectorAddr, err)
	}

	return conn, err
}

func initTracerProvider(
	ctx context.Context,
	res *resource.Resource,
	conn *grpc.ClientConn,
) (OtelCloser, error) {
	slog.Debug("setting up trace exporter")

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return traceExporter.Shutdown, nil
}

func initMeterProvider(
	ctx context.Context,
	res *resource.Resource,
	conn *grpc.ClientConn,
) (OtelCloser, error) {
	slog.Debug("setting up metric exporter")

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	return meterProvider.Shutdown, nil
}

func SetError(s trace.Span, err error) {
	s.RecordError(err)
	s.SetStatus(codes.Error, "Failure")
	s.SetAttributes(attribute.String("error.msg", err.Error()))
}
