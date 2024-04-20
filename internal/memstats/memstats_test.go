package memstats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemStatsForServer_GetMap(t *testing.T) {
	ms, err := NewMemStatsForServer()

	require.NoError(t, err)

	wantElements := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"TotalMemory",
		"FreeMemory",
		"CPUutilization1",
		"RandomValue",
	}

	t.Run("positive test", func(t *testing.T) {
		gotMapMS := ms.GetMap()

		for _, val := range wantElements {
			require.Contains(t, gotMapMS, val)
		}
	})
}
