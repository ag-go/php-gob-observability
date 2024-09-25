package statistics

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Boundary struct {
	name  string
	value float64
}

type StatisticSet struct {
	counters  map[string]uint64
	durations map[string]float64
	buckets   map[string]uint64
}

type Statistics struct {
	mutex      sync.Mutex
	names      map[string]StatisticSet
	boundaries []Boundary
}

func New(boundaries []float64) *Statistics {
	s := Statistics{
		names:      map[string]StatisticSet{},
		boundaries: []Boundary{},
	}
	sort.Float64s(boundaries)
	for _, value := range boundaries {
		name := fmt.Sprintf("%g", value)
		s.boundaries = append(s.boundaries, Boundary{name, value})
	}
	s.boundaries = append(s.boundaries, Boundary{"+Inf", math.MaxFloat64})
	return &s
}

func (s *Statistics) Add(name string, tagName string, tagValue string, duration float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := name + "|" + tagName
	ss, exists := s.names[key]
	if !exists {
		ss = StatisticSet{
			counters:  map[string]uint64{},
			durations: map[string]float64{},
			buckets:   map[string]uint64{},
		}
		s.names[key] = ss
	}
	ss.counters[tagValue]++
	ss.durations[tagValue] += duration
	for i := len(s.boundaries) - 1; i >= 0; i-- {
		b := s.boundaries[i]
		if b.value < duration {
			break
		}
		ss.buckets[b.name]++
	}
}

func (s *Statistics) Write(writer *http.ResponseWriter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	names := make([]string, 0, len(s.names))
	for name := range s.names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ss := s.names[name]
		parts := strings.SplitN(name, "|", 2)
		metricName := parts[0]
		tagName := parts[1]
		// counters
		(*writer).Write([]byte("# HELP " + metricName + "_seconds A summary of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
		(*writer).Write([]byte("# TYPE " + metricName + "_seconds summary\n"))
		keys := make([]string, 0, len(ss.counters))
		for key := range ss.counters {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		count := uint64(0)
		sum := float64(0)
		for _, k := range keys {
			c := ss.counters[k]
			count += c
			(*writer).Write([]byte(metricName + "_seconds_count{" + tagName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
			s := ss.durations[k]
			sum += s
			(*writer).Write([]byte(metricName + "_seconds_sum{" + tagName + "=" + strconv.Quote(k) + "} " + strconv.FormatFloat(s, 'f', 3, 64) + "\n"))
		}
		// totals
		(*writer).Write([]byte("# HELP " + metricName + "_total_seconds A histogram of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
		(*writer).Write([]byte("# TYPE " + metricName + "_total_seconds histogram\n"))
		for _, b := range s.boundaries {
			v := ss.buckets[b.name]
			(*writer).Write([]byte(metricName + "_total_seconds_bucket{le=" + strconv.Quote(b.name) + "} " + strconv.FormatUint(v, 10) + "\n"))
		}
		(*writer).Write([]byte(metricName + "_total_seconds_sum " + strconv.FormatFloat(sum, 'f', 3, 64) + "\n"))
		(*writer).Write([]byte(metricName + "_total_seconds_count " + strconv.FormatUint(count, 10) + "\n"))
		(*writer).Write([]byte("# EOF"))
	}
}
