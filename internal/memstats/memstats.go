// Package memstats defines structure for calculating memory statistic, virtual memory statistics, CPU utilization and random value.
package memstats

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// ForServer stores statistics data and defines methods for working with it.
type ForServer struct {
	*mem.VirtualMemoryStat
	runtime.MemStats
	sync.RWMutex
	CPUutilization float64
	RandomValue    float64
}

// NewForServer create new mem stats for server
func NewForServer() (*ForServer, error) {
	ms := &ForServer{}
	err := ms.ReadMemStats()
	if err != nil {
		return nil, fmt.Errorf("read mem stats: %w", err)
	}
	return ms, nil
}

// ReadMemStats do updates memory data.
func (ms *ForServer) ReadMemStats() error {
	ms.Lock()
	defer ms.Unlock()
	var err error

	runtime.ReadMemStats(&ms.MemStats)

	ms.VirtualMemoryStat, err = mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("get virtual memory: %w", err)
	}

	CPUutilization, err := cpu.Percent(0, false)
	if err != nil {
		return fmt.Errorf("get cpu unitilization")
	}

	ms.CPUutilization = CPUutilization[0]
	ms.RandomValue = rand.Float64()

	return nil
}

// GetMap return memory data.
func (ms *ForServer) GetMap() map[string]float64 {
	ms.RLock()
	defer ms.RUnlock()
	return map[string]float64{
		"Alloc":           float64(ms.Alloc),
		"BuckHashSys":     float64(ms.BuckHashSys),
		"Frees":           float64(ms.Frees),
		"GCCPUFraction":   ms.GCCPUFraction,
		"GCSys":           float64(ms.GCSys),
		"HeapAlloc":       float64(ms.HeapAlloc),
		"HeapIdle":        float64(ms.HeapIdle),
		"HeapInuse":       float64(ms.HeapInuse),
		"HeapObjects":     float64(ms.HeapObjects),
		"HeapReleased":    float64(ms.HeapReleased),
		"HeapSys":         float64(ms.HeapSys),
		"LastGC":          float64(ms.LastGC),
		"Lookups":         float64(ms.Lookups),
		"MCacheInuse":     float64(ms.MCacheInuse),
		"MCacheSys":       float64(ms.MCacheSys),
		"MSpanInuse":      float64(ms.MSpanInuse),
		"MSpanSys":        float64(ms.MSpanSys),
		"Mallocs":         float64(ms.Mallocs),
		"NextGC":          float64(ms.NextGC),
		"NumForcedGC":     float64(ms.NumForcedGC),
		"NumGC":           float64(ms.NumGC),
		"OtherSys":        float64(ms.OtherSys),
		"PauseTotalNs":    float64(ms.PauseTotalNs),
		"StackInuse":      float64(ms.StackInuse),
		"StackSys":        float64(ms.StackSys),
		"Sys":             float64(ms.Sys),
		"TotalAlloc":      float64(ms.TotalAlloc),
		"TotalMemory":     float64(ms.Total),
		"FreeMemory":      float64(ms.Free),
		"CPUutilization1": ms.CPUutilization,
		"RandomValue":     ms.RandomValue,
	}
}
