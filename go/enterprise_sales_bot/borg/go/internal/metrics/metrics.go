// Package metrics provides a typed metrics service with counters, gauges,
// histograms, and Prometheus export — ported from
// packages/core/src/services/MetricsService.ts.
//
// Thread-safe, singleton-capable, with event tracking, downsampled series,
// and system monitoring.
package metrics

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MetricEvent is a legacy event-stream record.
type MetricEvent struct {
	Timestamp int64             `json:"timestamp"`
	Type      string            `json:"type"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// HistogramStats holds statistical summaries for a histogram.
type HistogramStats struct {
	Count int     `json:"count"`
	Sum   float64 `json:"sum"`
	Avg   float64 `json:"avg"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
}

// AllMetrics holds a snapshot of all metric types.
type AllMetrics struct {
	Counters   map[string]float64        `json:"counters"`
	Gauges     map[string]float64        `json:"gauges"`
	Histograms map[string]HistogramStats `json:"histograms"`
}

// StatsResult holds the result of getStats().
type StatsResult struct {
	WindowMs    int64                    `json:"windowMs"`
	TotalEvents int                      `json:"totalEvents"`
	Counts      map[string]float64       `json:"counts"`
	Averages    map[string]float64       `json:"averages"`
	Counters    map[string]float64       `json:"counters"`
	Gauges      map[string]float64       `json:"gauges"`
	Histograms  map[string]HistogramStats `json:"histograms"`
	Series      []DownsampledBucket      `json:"series"`
}

// DownsampledBucket is a single bucket in a downsampled time series.
type DownsampledBucket struct {
	Time     int64   `json:"time"`
	Count    int     `json:"count"`
	ValueAvg float64 `json:"value_avg"`
}

// labeledValue stores a metric value with optional labels.
type labeledValue struct {
	value  float64
	labels map[string]string
}

// MetricsService provides typed metrics with counters, gauges, histograms.
type MetricsService struct {
	mu sync.RWMutex

	// Legacy event stream
	events    []MetricEvent
	maxEvents int

	// Typed metric storage
	counters      map[string]*labeledValue
	gauges        map[string]*labeledValue
	histograms    map[string][]float64

	// Monitoring
	monitorStop chan struct{}
	running     int32

	// Callback
	onMetric func(name string, mtype string, value float64, labels map[string]string)
}

var (
	globalInstance atomic.Pointer[MetricsService]
)

// GetMetricsService returns the global singleton MetricsService.
func GetMetricsService() *MetricsService {
	if inst := globalInstance.Load(); inst != nil {
		return inst
	}
	inst := &MetricsService{
		events:     make([]MetricEvent, 0, 10000),
		maxEvents:  10000,
		counters:   make(map[string]*labeledValue),
		gauges:     make(map[string]*labeledValue),
		histograms: make(map[string][]float64),
		monitorStop: make(chan struct{}),
	}
	if globalInstance.CompareAndSwap(nil, inst) {
		return inst
	}
	return globalInstance.Load()
}

// NewMetricsService creates a fresh, non-singleton MetricsService.
func NewMetricsService() *MetricsService {
	return &MetricsService{
		events:     make([]MetricEvent, 0, 10000),
		maxEvents:  10000,
		counters:   make(map[string]*labeledValue),
		gauges:     make(map[string]*labeledValue),
		histograms: make(map[string][]float64),
		monitorStop: make(chan struct{}),
	}
}

// --- Counters ---

func (ms *MetricsService) IncCounter(name string, value float64, labels map[string]string) {
	key := compositeKey(name, labels)
	ms.mu.Lock()
	lv := ms.counters[key]
	if lv == nil {
		lv = &labeledValue{labels: labels}
		ms.counters[key] = lv
	}
	lv.value += value
	ms.mu.Unlock()
	ms.emitCallback(name, "counter", lv.value, labels)
}

func (ms *MetricsService) GetCounter(name string, labels map[string]string) float64 {
	key := compositeKey(name, labels)
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if lv, ok := ms.counters[key]; ok {
		return lv.value
	}
	return 0
}

// --- Gauges ---

func (ms *MetricsService) SetGauge(name string, value float64, labels map[string]string) {
	key := compositeKey(name, labels)
	ms.mu.Lock()
	lv := ms.gauges[key]
	if lv == nil {
		lv = &labeledValue{labels: labels}
		ms.gauges[key] = lv
	}
	lv.value = value
	ms.mu.Unlock()
	ms.emitCallback(name, "gauge", value, labels)
}

func (ms *MetricsService) GetGauge(name string, labels map[string]string) float64 {
	key := compositeKey(name, labels)
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if lv, ok := ms.gauges[key]; ok {
		return lv.value
	}
	return 0
}

func (ms *MetricsService) IncGauge(name string, value float64, labels map[string]string) {
	key := compositeKey(name, labels)
	ms.mu.Lock()
	lv := ms.gauges[key]
	if lv == nil {
		lv = &labeledValue{labels: labels}
		ms.gauges[key] = lv
	}
	lv.value += value
	ms.mu.Unlock()
	ms.emitCallback(name, "gauge", lv.value, labels)
}

func (ms *MetricsService) DecGauge(name string, value float64, labels map[string]string) {
	ms.IncGauge(name, -value, labels)
}

// --- Histograms ---

func (ms *MetricsService) ObserveHistogram(name string, value float64) {
	ms.mu.Lock()
	ms.histograms[name] = append(ms.histograms[name], value)
	ms.mu.Unlock()
}

func (ms *MetricsService) GetHistogramStats(name string) *HistogramStats {
	ms.mu.RLock()
	values, ok := ms.histograms[name]
	ms.mu.RUnlock()
	if !ok || len(values) == 0 {
		return nil
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	count := len(sorted)
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}

	return &HistogramStats{
		Count: count,
		Sum:   sum,
		Avg:   sum / float64(count),
		Min:   sorted[0],
		Max:   sorted[count-1],
		P50:   sorted[int(float64(count)*0.5)],
		P95:   sorted[min(int(float64(count)*0.95), count-1)],
		P99:   sorted[min(int(float64(count)*0.99), count-1)],
	}
}

// Timer returns a function that, when called, records the elapsed duration.
func (ms *MetricsService) Timer(name string) func() {
	start := time.Now()
	return func() {
		ms.ObserveHistogram(name, float64(time.Since(start).Milliseconds()))
	}
}

// --- Export ---

func (ms *MetricsService) ExportPrometheus() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var lines []string

	// Counters
	counterNames := baseNames(ms.counters)
	for _, name := range counterNames {
		lines = append(lines, fmt.Sprintf("# TYPE %s counter", name))
		for key, lv := range ms.counters {
			base := keyBase(key)
			if base != name {
				continue
			}
			labelStr := formatLabels(lv.labels)
			if labelStr != "" {
				lines = append(lines, fmt.Sprintf("%s{%s} %.2f", name, labelStr, lv.value))
			} else {
				lines = append(lines, fmt.Sprintf("%s %.2f", name, lv.value))
			}
		}
	}

	// Gauges
	gaugeNames := baseNames(ms.gauges)
	for _, name := range gaugeNames {
		lines = append(lines, fmt.Sprintf("# TYPE %s gauge", name))
		for key, lv := range ms.gauges {
			base := keyBase(key)
			if base != name {
				continue
			}
			labelStr := formatLabels(lv.labels)
			if labelStr != "" {
				lines = append(lines, fmt.Sprintf("%s{%s} %.2f", name, labelStr, lv.value))
			} else {
				lines = append(lines, fmt.Sprintf("%s %.2f", name, lv.value))
			}
		}
	}

	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	return result
}

// GetAll returns a snapshot of all metrics.
func (ms *MetricsService) GetAll() *AllMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	counters := make(map[string]float64)
	for key, lv := range ms.counters {
		base := keyBase(key)
		counters[base] += lv.value
	}

	gauges := make(map[string]float64)
	for key, lv := range ms.gauges {
		base := keyBase(key)
		gauges[base] = lv.value
	}

	histograms := make(map[string]HistogramStats)
	for name := range ms.histograms {
		if stats := ms.GetHistogramStats(name); stats != nil {
			histograms[name] = *stats
		}
	}

	return &AllMetrics{Counters: counters, Gauges: gauges, Histograms: histograms}
}

// --- Legacy API ---

func (ms *MetricsService) Track(mtype string, value float64, tags map[string]string) {
	ms.mu.Lock()
	ms.events = append(ms.events, MetricEvent{
		Timestamp: time.Now().UnixMilli(),
		Type:      mtype,
		Value:     value,
		Tags:      tags,
	})
	if len(ms.events) > ms.maxEvents {
		ms.events = ms.events[len(ms.events)-ms.maxEvents/2:]
	}
	ms.mu.Unlock()

	// Bridge to typed metrics
	switch mtype {
	case "duration", "tool_execution":
		name := mtype
		if tags != nil {
			if t, ok := tags["tool"]; ok {
				name = t
			} else if t, ok := tags["name"]; ok {
				name = t
			}
		}
		ms.ObserveHistogram("duration_"+name, value)
	case "memory_heap", "memory_rss", "system_load", "system_free_mem":
		ms.SetGauge(mtype, value, tags)
	default:
		name := mtype
		if tags != nil {
			if t, ok := tags["tool"]; ok {
				name = t
			} else if t, ok := tags["name"]; ok {
				name = t
			}
		}
		ms.IncCounter(name, value, tags)
	}
}

func (ms *MetricsService) TrackDuration(name string, ms_ float64, tags map[string]string) {
	ms.Track("duration", ms_, mergeTags(tags, map[string]string{"name": name}))
}

// GetStats returns statistics for events within the given time window.
func (ms *MetricsService) GetStats(windowMs int64) *StatsResult {
	if windowMs <= 0 {
		windowMs = 3600000
	}
	now := time.Now().UnixMilli()
	cutoff := now - windowMs

	ms.mu.RLock()
	var relevant []MetricEvent
	for _, e := range ms.events {
		if e.Timestamp > cutoff {
			relevant = append(relevant, e)
		}
	}
	ms.mu.RUnlock()

	counts := make(map[string]float64)
	sums := make(map[string]float64)
	typeCounts := make(map[string]int)

	for _, e := range relevant {
		counts[e.Type] += e.Value
		sums[e.Type] += e.Value
		typeCounts[e.Type]++
	}

	averages := make(map[string]float64)
	for k, sum := range sums {
		averages[k] = sum / float64(typeCounts[k])
	}

	typed := ms.GetAll()

	return &StatsResult{
		WindowMs:    windowMs,
		TotalEvents: len(relevant),
		Counts:      counts,
		Averages:    averages,
		Counters:    typed.Counters,
		Gauges:      typed.Gauges,
		Histograms:  typed.Histograms,
		Series:      ms.downsample(relevant, 60),
	}
}

// StartMonitoring begins periodic system metrics collection.
func (ms *MetricsService) StartMonitoring(intervalMs int64) {
	if intervalMs <= 0 {
		intervalMs = 5000
	}
	if !atomic.CompareAndSwapInt32(&ms.running, 0, 1) {
		return // Already running
	}

	go func() {
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()

		var memStats runtime.MemStats
		for {
			select {
			case <-ticker.C:
				runtime.ReadMemStats(&memStats)
				ms.Track("memory_heap", float64(memStats.HeapAlloc), nil)
				ms.Track("memory_rss", float64(memStats.Sys), nil)
			case <-ms.monitorStop:
				return
			}
		}
	}()
}

// StopMonitoring stops the periodic monitoring goroutine.
func (ms *MetricsService) StopMonitoring() {
	if atomic.CompareAndSwapInt32(&ms.running, 1, 0) {
		close(ms.monitorStop)
		ms.monitorStop = make(chan struct{})
	}
}

// Reset clears all metrics.
func (ms *MetricsService) Reset() {
	ms.mu.Lock()
	ms.counters = make(map[string]*labeledValue)
	ms.gauges = make(map[string]*labeledValue)
	ms.histograms = make(map[string][]float64)
	ms.events = nil
	ms.mu.Unlock()
}

// OnMetric registers a callback for metric events.
func (ms *MetricsService) OnMetric(fn func(name string, mtype string, value float64, labels map[string]string)) {
	ms.onMetric = fn
}

func (ms *MetricsService) emitCallback(name, mtype string, value float64, labels map[string]string) {
	if ms.onMetric != nil {
		ms.onMetric(name, mtype, value, labels)
	}
}

// --- internal ---

func (ms *MetricsService) downsample(events []MetricEvent, buckets int) []DownsampledBucket {
	if len(events) == 0 {
		return nil
	}

	start := events[0].Timestamp
	end := time.Now().UnixMilli()
	interval := (end - start) / int64(buckets)
	if interval <= 0 {
		interval = 1
	}

	result := make([]DownsampledBucket, buckets)
	for i := 0; i < buckets; i++ {
		bucketStart := start + int64(i)*interval
		bucketEnd := bucketStart + interval

		var inBucket []MetricEvent
		for _, e := range events {
			if e.Timestamp >= bucketStart && e.Timestamp < bucketEnd {
				inBucket = append(inBucket, e)
			}
		}

		avg := 0.0
		if len(inBucket) > 0 {
			sum := 0.0
			for _, e := range inBucket {
				sum += e.Value
			}
			avg = sum / float64(len(inBucket))
		}

		result[i] = DownsampledBucket{
			Time:     bucketStart,
			Count:    len(inBucket),
			ValueAvg: avg,
		}
	}

	return result
}

func compositeKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	// Sorted key=value pairs
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, len(keys))
	for i, k := range keys {
		pairs[i] = k + "=" + labels[k]
	}
	return name + "{" + stringsJoin(pairs, ",") + "}"
}

func keyBase(key string) string {
	idx := indexByte(key, '{')
	if idx >= 0 {
		return key[:idx]
	}
	return key
}

func baseNames(m map[string]*labeledValue) []string {
	seen := make(map[string]bool)
	var names []string
	for key := range m {
		base := keyBase(key)
		if !seen[base] {
			seen[base] = true
			names = append(names, base)
		}
	}
	sort.Strings(names)
	return names
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf(`%s="%s"`, k, labels[k])
	}
	return stringsJoin(parts, ",")
}

func mergeTags(a, b map[string]string) map[string]string {
	result := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

func stringsJoin(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure math is used
var _ = math.Pi
