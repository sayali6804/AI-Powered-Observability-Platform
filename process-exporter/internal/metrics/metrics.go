package metrics

import "github.com/prometheus/client_golang/prometheus"

// Process metrics
var (
	processCPUPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_cpu_percent",
			Help: "CPU usage percentage for a specific process",
		},
		[]string{"process_name", "pid", "service"},
	)

	processMemoryWorkingSet = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_memory_working_set_bytes",
			Help: "Working set memory usage in bytes for a specific process",
		},
		[]string{"process_name", "pid", "service"},
	)

	processMemoryPrivate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_memory_private_bytes",
			Help: "Private memory usage in bytes for a specific process",
		},
		[]string{"process_name", "pid", "service"},
	)

	// System metrics
	totalCPUPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_cpu_percent",
			Help: "Total CPU usage percentage of the machine",
		},
	)

	totalMemoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_memory_usage_bytes",
			Help: "Total memory usage in bytes of the machine",
		},
	)
)

// RegisterMetrics registers all metrics with Prometheus
func RegisterMetrics() {
	prometheus.MustRegister(processCPUPercent)
	prometheus.MustRegister(processMemoryWorkingSet)
	prometheus.MustRegister(processMemoryPrivate)
	prometheus.MustRegister(totalCPUPercent)
	prometheus.MustRegister(totalMemoryUsage)
}