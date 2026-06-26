package collector

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jittakal/svx-cs-process-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/process"
)

// Register metrics
func init() {
	// Register process metrics
	prometheus.MustRegister(processCPUPercent)
	prometheus.MustRegister(processMemoryWorkingSet)
	prometheus.MustRegister(processMemoryPrivate)
}

// RegisterCollectors sets up and registers all collectors
func RegisterCollectors(processes []config.Process) {
	// Create process configs from the config.Process structs
	var processConfigs []ProcessConfig
	for _, p := range processes {
		processConfigs = append(processConfigs, ProcessConfig{
			Name:    p.Name,
			Pattern: p.Pattern,
		})
	}

	// In RegisterCollectors function:
	go func() {
		ticker := time.NewTicker(15 * time.Second) // Collect every 15 seconds
		defer ticker.Stop()

		for range ticker.C {
			CollectProcessMetrics(processConfigs)
			CollectSystemMetrics()
		}
	}()
}

// ProcessConfig holds the configuration for monitored processes
type ProcessConfig struct {
	Name    string
	Pattern string
}

func CollectProcessMetrics(processConfigs []ProcessConfig) {
	// Reset metrics before collecting
	processCPUPercent.Reset()
	processMemoryWorkingSet.Reset()
	processMemoryPrivate.Reset()

	// Create a map for quick pattern lookups
	patternMap := make(map[string]string) // map[processName]pattern
	for _, config := range processConfigs {
		pattern := config.Pattern
		if pattern == "" {
			pattern = config.Name
		}
		patternMap[config.Name] = strings.ToLower(pattern)
	}

	// Get processes directly by name using Windows API if possible
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error getting process list: %v", err)
		return
	}

	// Process cache to avoid duplicate metric collection for the same process name
	processCache := make(map[string]bool)

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		nameLower := strings.ToLower(name)

		// Check if this process matches any configured pattern using efficient lookups
		for procName, pattern := range patternMap {
			// Skip if we already found this process
			cacheKey := procName + "|" + strconv.Itoa(int(p.Pid))
			if _, exists := processCache[cacheKey]; exists {
				continue
			}

			if strings.Contains(nameLower, pattern) {
				pid := p.Pid
				pidStr := strconv.Itoa(int(pid))
				log.Printf("Collecting metrics for process: %s (PID: %s)", procName, pidStr)

				// Mark as processed
				processCache[cacheKey] = true

				// Collect metrics asynchronously
				go func(p *process.Process, procName, pidStr string) {
					// Collect CPU usage with timeout
					cpuChan := make(chan float64, 1)
					errChan := make(chan error, 1)

					go func() {
						cpuPercent, err := p.CPUPercent()
						if err != nil {
							errChan <- err
							return
						}
						cpuChan <- cpuPercent
					}()

					// Wait for CPU metric with timeout
					select {
					case cpuPercent := <-cpuChan:
						processCPUPercent.WithLabelValues(procName, pidStr).Set(cpuPercent)
					case err := <-errChan:
						log.Printf("Error collecting CPU for %s (PID %s): %v", procName, pidStr, err)
					case <-time.After(500 * time.Millisecond):
						log.Printf("Timeout collecting CPU for %s (PID %s)", procName, pidStr)
					}

					// Collect memory metrics (these are usually faster)
					memInfo, err := p.MemoryInfo()
					if err == nil {
						processMemoryWorkingSet.WithLabelValues(procName, pidStr).Set(float64(memInfo.RSS))
						processMemoryPrivate.WithLabelValues(procName, pidStr).Set(float64(memInfo.VMS))
					}
				}(p, procName, pidStr)

				break // Found a match for this process
			}
		}
	}
}

// Updated to match the pattern field from config.Process
func CollectProcessMetricsOld(processConfigs []ProcessConfig) {
	// Reset metrics before collecting
	processCPUPercent.Reset()
	processMemoryWorkingSet.Reset()
	processMemoryPrivate.Reset()

	// Get all processes
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error getting process list: %v", err)
		return
	}

	// Check each process
	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue // Skip if we can't get process name
		}

		// fmt.Println("Name of the process:", name)

		// Check if this process matches any of our configured processes
		for _, config := range processConfigs {
			pattern := config.Pattern
			if pattern == "" {
				pattern = config.Name // Use name as pattern if pattern is not specified
			}

			if strings.Contains(strings.ToLower(name), strings.ToLower(pattern)) {
				pid := p.Pid
				pidStr := strconv.Itoa(int(pid))

				log.Println("cpu and memory for process:", config.Name, "pid:", pidStr)

				// Collect CPU usage
				cpuPercent, err := p.CPUPercent()
				if err == nil {
					processCPUPercent.WithLabelValues(config.Name, pidStr).Set(cpuPercent)
				}

				// Collect memory metrics
				memInfo, err := p.MemoryInfo()
				if err == nil {
					processMemoryWorkingSet.WithLabelValues(config.Name, pidStr).Set(float64(memInfo.RSS))
					processMemoryPrivate.WithLabelValues(config.Name, pidStr).Set(float64(memInfo.VMS))
				}

				// Break out of the loop once we find a match
				break
			}
		}
	}
}
