package perf

import (
	"sort"
	"time"

	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

type average struct {
	// nanoseconds
	total int64
	// recordings
	count int64
}

var enabled bool
var averages = make(map[string]average)

func RecordAverageTime(key string, nanos int64) {
	if !enabled {
		return
	}

	var avg average
	if v, ok := averages[key]; ok {
		avg = v
	}

	avg.total += nanos
	avg.count++
	averages[key] = avg
}

func SetMetricsEnabled(enable bool) {
	enabled = enable
}

func LogMetrics() {
	if !enabled || len(averages) == 0 {
		return
	}

	keys := make([]string, 0, len(averages))
	for k := range averages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	log.Perf("average metrics")
	for _, k := range keys {
		if v, ok := averages[k]; ok {
			avg := time.Duration(v.total / v.count)
			log.Perff("- %v = %v", k, avg)
		}
	}
}
