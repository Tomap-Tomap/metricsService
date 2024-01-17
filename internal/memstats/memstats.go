package memstats

import (
	"runtime"
	"strconv"
)

type StringMS struct {
	Name, Value string
}

func GetMemStatsForServer(ms *runtime.MemStats) []StringMS {
	return []StringMS{
		{"Alloc", strconv.FormatUint(ms.Alloc, 10)},
		{"BuckHashSys", strconv.FormatUint(ms.BuckHashSys, 10)},
		{"Frees", strconv.FormatUint(ms.Frees, 10)},
		{"GCCPUFraction", strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)},
		{"GCSys", strconv.FormatUint(ms.GCSys, 10)},
		{"HeapAlloc", strconv.FormatUint(ms.HeapAlloc, 10)},
		{"HeapIdle", strconv.FormatUint(ms.HeapIdle, 10)},
		{"HeapInuse", strconv.FormatUint(ms.HeapInuse, 10)},
		{"HeapObjects", strconv.FormatUint(ms.HeapObjects, 10)},
		{"HeapReleased", strconv.FormatUint(ms.HeapReleased, 10)},
		{"HeapSys", strconv.FormatUint(ms.HeapSys, 10)},
		{"LastGC", strconv.FormatUint(ms.LastGC, 10)},
		{"Lookups", strconv.FormatUint(ms.Lookups, 10)},
		{"MCacheInuse", strconv.FormatUint(ms.MCacheInuse, 10)},
		{"MCacheSys", strconv.FormatUint(ms.MCacheSys, 10)},
		{"MSpanInuse", strconv.FormatUint(ms.MSpanInuse, 10)},
		{"MSpanSys", strconv.FormatUint(ms.MSpanSys, 10)},
		{"Mallocs", strconv.FormatUint(ms.Mallocs, 10)},
		{"NextGC", strconv.FormatUint(ms.NextGC, 10)},
		{"NumForcedGC", strconv.FormatUint(uint64(ms.NumForcedGC), 10)},
		{"NumGC", strconv.FormatUint(uint64(ms.NumGC), 10)},
		{"OtherSys", strconv.FormatUint(ms.OtherSys, 10)},
		{"PauseTotalNs", strconv.FormatUint(ms.PauseTotalNs, 10)},
		{"StackInuse", strconv.FormatUint(ms.StackInuse, 10)},
		{"StackSys", strconv.FormatUint(ms.StackSys, 10)},
		{"Sys", strconv.FormatUint(ms.Sys, 10)},
		{"TotalAlloc", strconv.FormatUint(ms.TotalAlloc, 10)},
	}
}
