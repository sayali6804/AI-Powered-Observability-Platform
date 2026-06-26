package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Process metrics
var (
	processCPUPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_cpu_percent",
			Help: "CPU usage percentage for a specific process",
		},
		[]string{"process_name", "pid"},
	)

	processMemoryWorkingSet = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_memory_working_set_bytes",
			Help: "Working set memory usage in bytes",
		},
		[]string{"process_name", "pid"},
	)

	processMemoryPrivate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "process_memory_private_bytes",
			Help: "Private memory usage in bytes",
		},
		[]string{"process_name", "pid"},
	)
)
