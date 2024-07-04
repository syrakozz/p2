// Package stress tests API limits.
package stress

import (
	"disruptive/lib/common"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// Results contains stat values.
type Results struct {
	sync.Mutex
	Num       int             `json:"num"`
	Total     time.Duration   `json:"-"`
	TotalS    string          `json:"total"`
	Durations []time.Duration `json:"-"`
	Stats     struct {
		Num   int           `json:"num"`
		Min   time.Duration `json:"-"`
		MinS  string        `json:"min"`
		Max   time.Duration `json:"-"`
		MaxS  string        `json:"max"`
		Avg   time.Duration `json:"-"`
		AvgS  string        `json:"avg"`
		Mode  time.Duration `json:"-"`
		ModeS string        `json:"mode"`
	} `json:"stats"`
}

var (
	results Results
)

func printResults(results *Results) {
	results.Stats.Min = math.MaxInt64

	sortDurations(results.Durations)

	for _, d := range results.Durations {
		if d < results.Stats.Min {
			results.Stats.Min = d
		}

		if d > results.Stats.Max {
			results.Stats.Max = d
		}

		results.Stats.Avg += d
	}

	l := len(results.Durations)
	if l == 0 {
		return
	}

	results.Stats.Num = l
	results.Stats.Avg = time.Duration(results.Stats.Avg.Nanoseconds() / int64(l))

	results.Stats.MinS = results.Stats.Min.String()
	results.Stats.MaxS = results.Stats.Max.String()
	results.Stats.AvgS = results.Stats.Avg.String()

	results.Stats.Mode = results.Durations[l/2]
	results.Stats.ModeS = results.Stats.Mode.String()

	fmt.Println(common.MarshalIndent(results))
}

func sortDurations(durations []time.Duration) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
}
