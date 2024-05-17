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
	// var virtualMemoryTests = []struct {
	// 	mockedRootFS string
	// 	stat         *mem.VirtualMemoryStat
	// }{
	// 	{
	// 		"intelcorei5", &mem.VirtualMemoryStat{
	// 			Total:          16502300672,
	// 			Available:      11495358464,
	// 			Used:           3437277184,
	// 			UsedPercent:    20.82907863769651,
	// 			Free:           8783491072,
	// 			Active:         4347392000,
	// 			Inactive:       2938834944,
	// 			Wired:          0,
	// 			Laundry:        0,
	// 			Buffers:        212496384,
	// 			Cached:         4069036032,
	// 			WriteBack:      0,
	// 			Dirty:          176128,
	// 			WriteBackTmp:   0,
	// 			Shared:         1222402048,
	// 			Slab:           253771776,
	// 			Sreclaimable:   186470400,
	// 			Sunreclaim:     67301376,
	// 			PageTables:     65241088,
	// 			SwapCached:     0,
	// 			CommitLimit:    16509730816,
	// 			CommittedAS:    12360818688,
	// 			HighTotal:      0,
	// 			HighFree:       0,
	// 			LowTotal:       0,
	// 			LowFree:        0,
	// 			SwapTotal:      8258580480,
	// 			SwapFree:       8258580480,
	// 			Mapped:         1172627456,
	// 			VmallocTotal:   35184372087808,
	// 			VmallocUsed:    0,
	// 			VmallocChunk:   0,
	// 			HugePagesTotal: 0,
	// 			HugePagesFree:  0,
	// 			HugePagesRsvd:  0,
	// 			HugePagesSurp:  0,
	// 			HugePageSize:   2097152,
	// 		},
	// 	},
	// 	{
	// 		"issue1002", &mem.VirtualMemoryStat{
	// 			Total:          260579328,
	// 			Available:      215199744,
	// 			Used:           34328576,
	// 			UsedPercent:    13.173944481121694,
	// 			Free:           124506112,
	// 			Active:         108785664,
	// 			Inactive:       8581120,
	// 			Wired:          0,
	// 			Laundry:        0,
	// 			Buffers:        4915200,
	// 			Cached:         96829440,
	// 			WriteBack:      0,
	// 			Dirty:          0,
	// 			WriteBackTmp:   0,
	// 			Shared:         0,
	// 			Slab:           9293824,
	// 			Sreclaimable:   2764800,
	// 			Sunreclaim:     6529024,
	// 			PageTables:     405504,
	// 			SwapCached:     0,
	// 			CommitLimit:    130289664,
	// 			CommittedAS:    25567232,
	// 			HighTotal:      134217728,
	// 			HighFree:       67784704,
	// 			LowTotal:       126361600,
	// 			LowFree:        56721408,
	// 			SwapTotal:      0,
	// 			SwapFree:       0,
	// 			Mapped:         38793216,
	// 			VmallocTotal:   1996488704,
	// 			VmallocUsed:    0,
	// 			VmallocChunk:   0,
	// 			HugePagesTotal: 0,
	// 			HugePagesFree:  0,
	// 			HugePagesRsvd:  0,
	// 			HugePagesSurp:  0,
	// 			HugePageSize:   0,
	// 		},
	// 	},
	// 	{
	// 		"anonhugepages", &mem.VirtualMemoryStat{
	// 			Total:         260799420 * 1024,
	// 			Available:     127880216 * 1024,
	// 			Free:          119443248 * 1024,
	// 			AnonHugePages: 50409472 * 1024,
	// 			Used:          144748720128,
	// 			UsedPercent:   54.20110673559013,
	// 		},
	// 	},
	// }

	// for _, tt := range virtualMemoryTests {
	// 	t.Run(tt.mockedRootFS, func(t *testing.T) {
	// 		t.Setenv("HOST_PROC", filepath.Join("testdata/linux/virtualmemory/", tt.mockedRootFS, "proc"))

	// 		stat, err := VirtualMemory()
	// 		skipIfNotImplementedErr(t, err)
	// 		if err != nil {
	// 			t.Errorf("error %v", err)
	// 		}
	// 		if !reflect.DeepEqual(stat, tt.stat) {
	// 			t.Errorf("got: %+v\nwant: %+v", stat, tt.stat)
	// 		}
	// 	})
	// }

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
