package metrics

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Bucket struct {
	Name  string
	Value float64
}

type MetricSet struct {
	Counters       map[string]uint64
	DurationCounts map[string]uint64
	DurationSums   map[string]float64
	Buckets        map[string]uint64
}

type Metrics struct {
	mutex   sync.Mutex
	Names   map[string]MetricSet
	Buckets []Bucket
}

func New() *Metrics {
	return NewWithBuckets([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})
}

func NewWithBuckets(buckets []float64) *Metrics {
	s := Metrics{
		Names:   map[string]MetricSet{},
		Buckets: []Bucket{},
	}
	sort.Float64s(buckets)
	for _, value := range buckets {
		name := fmt.Sprintf("%g", value)
		s.Buckets = append(s.Buckets, Bucket{name, value})
	}
	return &s
}

func (s *Metrics) Inc(name string, labelName string, labelValue string, delta uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := name + "|" + labelName
	ss, exists := s.Names[key]
	if !exists {
		ss = MetricSet{
			Counters:       map[string]uint64{},
			DurationCounts: map[string]uint64{},
			DurationSums:   map[string]float64{},
			Buckets:        map[string]uint64{},
		}
		s.Names[key] = ss
	}
	ss.Counters[labelValue] += delta
}

func (s *Metrics) Add(name string, labelName string, labelValue string, duration float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := name + "|" + labelName
	ss, exists := s.Names[key]
	if !exists {
		ss = MetricSet{
			Counters:       map[string]uint64{},
			DurationCounts: map[string]uint64{},
			DurationSums:   map[string]float64{},
			Buckets:        map[string]uint64{},
		}
		s.Names[key] = ss
	}
	ss.DurationCounts[labelValue]++
	ss.DurationSums[labelValue] += duration
	for i := len(s.Buckets) - 1; i >= 0; i-- {
		b := s.Buckets[i]
		if b.Value < duration {
			break
		}
		ss.Buckets[b.Name]++
	}
}

func (s *Metrics) Write(writer *http.ResponseWriter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	(*writer).Header().Set("Content-Encoding", "gzip")
	gw := gzip.NewWriter((*writer))
	defer gw.Close()
	names := make([]string, 0, len(s.Names))
	for name := range s.Names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ss := s.Names[name]
		parts := strings.SplitN(name, "|", 2)
		metricName := parts[0]
		labelName := parts[1]
		// counters
		if len(ss.Counters) > 0 {
			gw.Write([]byte("# TYPE " + metricName + " counter\n"))
			gw.Write([]byte("# HELP " + metricName + " A counter of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
			keys := make([]string, 0, len(ss.Counters))
			for key := range ss.Counters {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, k := range keys {
				c := ss.Counters[k]
				gw.Write([]byte(metricName + "_total{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
			}
		}
		// DurationCounts
		if len(ss.DurationCounts) > 0 {
			gw.Write([]byte("# TYPE " + metricName + "_seconds summary\n"))
			gw.Write([]byte("# UNIT " + metricName + "_seconds seconds\n"))
			gw.Write([]byte("# HELP " + metricName + "_seconds A summary of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
			keys := make([]string, 0, len(ss.DurationCounts))
			for key := range ss.DurationCounts {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			count := uint64(0)
			sum := float64(0)
			for _, k := range keys {
				c := ss.DurationCounts[k]
				count += c
				gw.Write([]byte(metricName + "_seconds_count{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
				s := ss.DurationSums[k]
				sum += s
				gw.Write([]byte(metricName + "_seconds_sum{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatFloat(s, 'f', 3, 64) + "\n"))
			}
			// totals
			gw.Write([]byte("# TYPE " + metricName + "_total_seconds histogram\n"))
			gw.Write([]byte("# UNIT " + metricName + "_total_seconds seconds\n"))
			gw.Write([]byte("# HELP " + metricName + "_total_seconds A histogram of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
			for _, b := range s.Buckets {
				v := ss.Buckets[b.Name]
				gw.Write([]byte(metricName + "_total_seconds_bucket{le=" + strconv.Quote(b.Name) + "} " + strconv.FormatUint(v, 10) + "\n"))
			}
			gw.Write([]byte(metricName + "_total_seconds_bucket{le=\"+Inf\"} " + strconv.FormatUint(count, 10) + "\n"))
			gw.Write([]byte(metricName + "_total_seconds_sum " + strconv.FormatFloat(sum, 'f', 3, 64) + "\n"))
			gw.Write([]byte(metricName + "_total_seconds_count " + strconv.FormatUint(count, 10) + "\n"))
		}
	}
	gw.Write([]byte("# EOF\n"))
}

func (s *Metrics) AddMetrics(s2 *Metrics) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for key, value := range s2.Names {
		ss, exists := s.Names[key]
		if !exists {
			s.Names[key] = value
		} else {
			for k, v := range value.Counters {
				ss.Counters[k] += v
			}
			for k, v := range value.DurationCounts {
				ss.DurationCounts[k] += v
			}
			for k, v := range value.DurationSums {
				ss.DurationSums[k] += v
			}
			for k, v := range value.Buckets {
				ss.Buckets[k] += v
			}
		}
	}
}

func (s *Metrics) WriteGob(writer *http.ResponseWriter) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return gob.NewEncoder((*writer)).Encode(s)
}

func (s *Metrics) ReadGob(resp *http.Response) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return gob.NewDecoder(resp.Body).Decode(s)
}
