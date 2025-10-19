package victoria_metrics

import (
	"context"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/rs/zerolog/log"
)

type Writer struct {
	Endpoint string
	Headers  map[string]string
	Labels   string
	Interval time.Duration
}

func (vm *Writer) Start(ctx context.Context, set *metrics.Set) error {
	var headers []string
	for k := range vm.Headers {
		headers = append(headers, k+": "+vm.Headers[k])
	}
	log.Info().Msgf("Sending last deal prices into %s with labels %s every %s",
		vm.Endpoint, vm.Labels, vm.Interval.String())
	return set.InitPushWithOptions(ctx, vm.Endpoint, vm.Interval, &metrics.PushOptions{
		ExtraLabels: vm.Labels,
		Headers:     headers,
	})
}

func (vm *Writer) StartSendingRuntimeMetrics(ctx context.Context) error {
	var headers []string
	for k := range vm.Headers {
		headers = append(headers, k+": "+vm.Headers[k])
	}
	log.Info().Msgf("Sending runtime metrics into %s with labels %s every %s",
		vm.Endpoint, vm.Labels, vm.Interval.String())
	return metrics.InitPushWithOptions(ctx, vm.Endpoint, vm.Interval, true, &metrics.PushOptions{
		ExtraLabels: vm.Labels,
		Headers:     headers,
	})
}
