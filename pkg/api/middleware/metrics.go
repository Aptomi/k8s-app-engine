package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type prometheusHandler struct {
	handler      http.Handler
	requests     *prometheus.CounterVec
	duration     *prometheus.HistogramVec
	responseSize *prometheus.HistogramVec
}

// NewMetricsHandler returns middleware that collects HTTP req/resp specific metrics
func NewMetricsHandler(serviceName string, handler http.Handler) http.Handler {
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Number of processed HTTP requests labeled with status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(requests)

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_request_duration_seconds",
		Help:        "Duration of the HTTP request processing labeled with status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     []float64{.01, .05, .1, .5, 1, 2.5, 5, 10, 20, 30, 50},
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(duration)

	responseSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_response_size_bytes",
		Help:        "Size of the HTTP response labeled with status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     prometheus.ExponentialBuckets(100, 10, 5),
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(responseSize)

	return &prometheusHandler{
		handler:      handler,
		requests:     requests,
		duration:     duration,
		responseSize: responseSize,
	}
}

func (h *prometheusHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()
	infoWriter := wrapResponseWrite(writer)

	defer func() {
		h.requests.WithLabelValues(http.StatusText(infoWriter.status), request.Method, request.URL.Path).Inc()
		h.duration.WithLabelValues(http.StatusText(infoWriter.status), request.Method, request.URL.Path).Observe(time.Since(start).Seconds())
		h.responseSize.WithLabelValues(http.StatusText(infoWriter.status), request.Method, request.URL.Path).Observe(float64(infoWriter.size))
	}()

	h.handler.ServeHTTP(infoWriter, request)
}

type infoResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func wrapResponseWrite(writer http.ResponseWriter) *infoResponseWriter {
	return &infoResponseWriter{
		ResponseWriter: writer,
		status:         http.StatusOK,
	}
}

func (w *infoResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size

	return size, err
}

func (w *infoResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}
