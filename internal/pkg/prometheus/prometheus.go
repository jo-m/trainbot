// Package prometheus configures, initializes and serves global application prometheus metrics.
package prometheus

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Init starts an HTTP server for the prometheus endpoint in the background.
func Init(prometheusListen string) {
	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              prometheusListen,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

// RecordFrameDisposition counts in which way a frame was used, or discarded.
func RecordFrameDisposition(disposition string) {
	frameDispositions.WithLabelValues(disposition).Inc()
}

// RecordSequenceLength sets the number of frames stored in memory for the current train.
func RecordSequenceLength(length int) {
	sequenceLength.Set(float64(length))
}

// RecordFitAndStitchResult counts fitAndStitch() successes and failure modes.
func RecordFitAndStitchResult(result string) {
	fitAndStitchResult.WithLabelValues(result).Inc()
}

// RecordBrightnessContrast counts brightness and contrast stats used for discarding bad frames.
func RecordBrightnessContrast(avg float64, avgDev float64) {
	brightnessAvg.Observe(avg)
	brightnessAvgDev.Observe(avgDev)
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
	fitAndStitchResult = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trainbot_fit_and_stitch_results_total",
			Help: "Results from fitAndStitch(). Eg. train detected, unable to fit.",
		},
		[]string{"result"},
	)
	brightnessAvg = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "trainbot_brightness_avg",
			Buckets: prometheus.ExponentialBucketsRange(0.0005, 1.0, 20),
		},
	)
	brightnessAvgDev = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "trainbot_brightness_avgdev",
			Buckets: prometheus.ExponentialBucketsRange(0.0005, 1.0, 20),
		},
	)
)
