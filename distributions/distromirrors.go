package distributions

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/thanoskoutr/gomirror/mirrors"
)

type DistributionMirrors struct {
	Distribution Distributor       `json:"distribution"`
	Mirrors      []*mirrors.Mirror `json:"urls"`
}

func (d DistributionMirrors) String() string {
	return fmt.Sprintf("{Distribution: %v, Mirrors: %v}", d.Distribution.Name(), d.Mirrors)
}

func (d *DistributionMirrors) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Distribution string            `json:"distribution"`
		Mirrors      []*mirrors.Mirror `json:"urls"`
	}{
		Distribution: d.Distribution.Name(),
		Mirrors:      d.Mirrors,
	})
}

func (d *DistributionMirrors) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	var err error
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v["distribution"] != nil {
		d.Distribution, err = ToDistribution(v["distribution"].(string))
		if err != nil {
			return err
		}
	}
	// if v["urls"] != nil {}

	return nil
}

// Replaces Mirror list (don't care if empty or not)
func (d *DistributionMirrors) UpdateMirrors(source mirrors.MirrorSource, filename string) {
	mirrorsList := d.Distribution.GetMirrors(source, filename)
	d.Mirrors = make([]*mirrors.Mirror, len(mirrorsList))
	// Create shallow copy and assign to slice of pointers
	for i, mirror := range mirrorsList {
		v := mirror
		d.Mirrors[i] = &v
	}
}

// TODO: Filter by host (1 host -> multiple mirrors (protocols))
// TODO: Customize progress bar
func (d *DistributionMirrors) UpdateMirrorStatistics(rounds int64) {
	// Keep statistics from all rounds
	allTimes := make([][]time.Duration, rounds)
	for round := range allTimes {
		allTimes[round] = make([]time.Duration, len(d.Mirrors))
	}

	// Run for each round
	for round := int64(0); round < rounds; round++ {
		fmt.Fprintf(os.Stderr, "Round %v\n", round)
		// Create wait group for all goroutines
		var wg sync.WaitGroup
		wg.Add(len(d.Mirrors))

		// Create progress bar
		bar := progressbar.Default(int64(len(d.Mirrors)), "Mirrors")

		// For each mirror create a goroutine for parallel requests
		for i, mirror := range d.Mirrors {
			if mirror.Statistics == nil {
				mirror.Statistics = &mirrors.MirrorStatistics{}
			}
			// Make request and update mirror statistics
			go func(round int64, i int, mirror *mirrors.Mirror) {
				totalTime := mirror.GetTime()
				// fmt.Fprintf(os.Stderr, "Mirror %v: Country: %v, URL: %v, Time: %v\n", i, mirror.Country, mirror.URL, totalTime)
				d.Mirrors[i].Statistics.ResponseTimeHTTP = totalTime
				allTimes[round][i] = totalTime
				wg.Done()
				bar.Add(1)
			}(round, i, mirror)
		}
		wg.Wait()
	}

	// Calculate averages
	accum := make([]time.Duration, len(d.Mirrors))
	for i := range allTimes[0] {
		for round := range allTimes {
			accum[i] += allTimes[round][i]
		}
		d.Mirrors[i].Statistics.AvgResponseTimeHTTP = time.Duration(int64(accum[i]) / rounds)
		// fmt.Fprintf(os.Stderr, "Mirror %v: Country: %v, URL: %v, Time: %v, Avg Time: %v, Accum Time: %v\n", i, d.Mirrors[i].Country, d.Mirrors[i].URL, d.Mirrors[i].Statistics.ResponseTimeHTTP, d.Mirrors[i].Statistics.AvgResponseTimeHTTP, accum[i])
	}

}

func (d DistributionMirrors) Len() int {
	return len(d.Mirrors)
}

func (d DistributionMirrors) Less(i, j int) bool {
	return d.Mirrors[i].Statistics.AvgResponseTimeHTTP < d.Mirrors[j].Statistics.AvgResponseTimeHTTP
}

func (d DistributionMirrors) Swap(i, j int) {
	d.Mirrors[i], d.Mirrors[j] = d.Mirrors[j], d.Mirrors[i]
}

// TODO: Sort based on other factors
func (d *DistributionMirrors) SortMirrors() {
	sort.Sort(d)
}

// TODO: Find based on other factors
func (d *DistributionMirrors) BestMirror() mirrors.Mirror {
	bestTime := time.Duration(math.MaxInt64)
	bestMirror := mirrors.Mirror{}
	for _, mirror := range d.Mirrors {
		if mirror.Statistics == nil {
			continue
		}
		if mirror.Statistics.AvgResponseTimeHTTP < bestTime {
			bestTime = mirror.Statistics.AvgResponseTimeHTTP
			bestMirror = *mirror
		}
	}
	return bestMirror
}
