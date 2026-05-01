package runtime

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	InputRecords  prometheus.Counter
	OutputRecords prometheus.Counter
	Errors        *prometheus.CounterVec
	Latency       prometheus.Histogram
}

func NewMetrics(component string) *Metrics {
	return &Metrics{
		InputRecords: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "kinetix_input_records_total",
			Help:        "Number of input records read by the component.",
			ConstLabels: prometheus.Labels{"component": component},
		}),
		OutputRecords: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "kinetix_output_records_total",
			Help:        "Number of output records written by the component.",
			ConstLabels: prometheus.Labels{"component": component},
		}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "kinetix_processing_errors_total",
			Help:        "Number of processing errors by category.",
			ConstLabels: prometheus.Labels{"component": component},
		}, []string{"category"}),
		Latency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:        "kinetix_processing_latency_seconds",
			Help:        "Time spent processing one record.",
			ConstLabels: prometheus.Labels{"component": component},
			Buckets:     prometheus.DefBuckets,
		}),
	}
}

func (m *Metrics) Register(reg *prometheus.Registry) {
	reg.MustRegister(m.InputRecords, m.OutputRecords, m.Errors, m.Latency)
}

func Serve(addr string, reg *prometheus.Registry) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	return server
}
