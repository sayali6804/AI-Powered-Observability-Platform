package collector

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var (
	totalCPUUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Total CPU usage percentage of the machine",
	})

	totalMemoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage_bytes",
		Help: "Total memory usage in bytes of the machine",
	})
)

func init() {
	prometheus.MustRegister(totalCPUUsage)
	prometheus.MustRegister(totalMemoryUsage)
}

// CollectSystemMetrics collects total CPU and memory usage metrics for the machine
func CollectSystemMetrics() {
	cpuUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		return
	}

	if len(cpuUsage) > 0 {
		totalCPUUsage.Set(cpuUsage[0])
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory info: %v", err)
		return
	}
	totalMemoryUsage.Set(float64(memInfo.Used))
}
