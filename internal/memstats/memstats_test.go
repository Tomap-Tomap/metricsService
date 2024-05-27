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

func BenchmarkReadMemStats(b *testing.B) {
	ms, _ := NewMemStatsForServer()

	for i := 0; i < b.N; i++ {
		ms.ReadMemStats()
	}
}

func BenchmarkGetMap(b *testing.B) {
	ms, _ := NewMemStatsForServer()

	for i := 0; i < b.N; i++ {
		ms.GetMap()
	}
}

func TestMemStatsForServer_ReadMemStats(t *testing.T) {
	t.Run("error read VM", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/error_meminfo")
		_, err := NewMemStatsForServer()

		require.Error(t, err)
	})

	t.Run("error read cpu", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/")
		_, err := NewMemStatsForServer()

		require.Error(t, err)
	})
}
