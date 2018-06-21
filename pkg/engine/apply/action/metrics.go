package action

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	mActionCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "aptomi_actions_total",
			Help:        "Number of processed actions labeled with kind.",
			ConstLabels: prometheus.Labels{"service": "aptomi"},
		},
		[]string{"kind", "name", "success"},
	)

	mActionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "aptomi_action_duration_seconds",
			Help:        "Duration of the processed action labeled with kind.",
			ConstLabels: prometheus.Labels{"service": "aptomi"},
			Buckets:     []float64{.01, .05, .1, .5, 1, 2.5, 5, 10, 20, 30, 50},
		},
		[]string{"kind", "name", "success"},
	)
)

func init() {
	prometheus.MustRegister(mActionCount)
	prometheus.MustRegister(mActionDuration)

}

// CollectMetricsFor collects metrics for the given action, start time and resulting error
func CollectMetricsFor(action Interface, start time.Time, err error) {
	labels := []string{action.GetKind(), action.GetName(), strconv.FormatBool(err == nil)}

	mActionCount.WithLabelValues(labels...).Inc()
	mActionDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
}
