// Package prometheus configures, initializes and serves global application prometheus metrics.
package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Init starts an HTTP server for the prometheus endpoint in the background.
func Init(prometheusListen string) {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(prometheusListen, nil)
}

// RecordFrameDisposition counts in which way a frame was used, or discarded.
func RecordFrameDisposition(disposition string) {
	frameDispositions.WithLabelValues(disposition).Inc()
}

// RecordSequenceLength sets the number of frames stored in memory for the current train.
func RecordSequenceLength(length int) {
	sequenceLength.Set(float64(length))
}

var (
	frameDispositions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trainbot_frame_dispositions_total",
			Help: "How frames were used",
		},
		[]string{"disposition"},
	)
	sequenceLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "trainbot_sequence_length",
			Help: "Current number of frames stored.",
		},
	)
)
