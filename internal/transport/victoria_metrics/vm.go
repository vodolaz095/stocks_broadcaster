package victoria_metrics

import (
	"context"
	"time"

	"github.com/VictoriaMetrics/metrics"
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
	return metrics.InitPushWithOptions(ctx, vm.Endpoint, vm.Interval, true, &metrics.PushOptions{
		ExtraLabels: vm.Labels,
		Headers:     headers,
	})
}
