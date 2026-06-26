package hsync

import (
	"context"
	"sync"
	"time"
)

type LinkCrawlerManager struct {
	mu             sync.RWMutex
	running        bool
	interval       time.Duration
	dbPath         string
	classifyTags   bool
	cancel         context.CancelFunc
	lastStartedAt  time.Time
	lastFinishedAt time.Time
	lastError      string
	lastReport     *LinkCrawlerReport
	totalRuns      int
	totalProcessed int
	totalSucceeded int
	totalFailed    int
	totalTagged    int
}

type LinkCrawlerStatus struct {
	Running        bool               `json:"running"`
	IntervalMs     int64              `json:"intervalMs"`
	ClassifyTags   bool               `json:"classifyTags"`
	DBPath         string             `json:"dbPath"`
	LastStartedAt  string             `json:"lastStartedAt,omitempty"`
	LastFinishedAt string             `json:"lastFinishedAt,omitempty"`
	LastError      string             `json:"lastError,omitempty"`
	LastReport     *LinkCrawlerReport `json:"lastReport,omitempty"`
	TotalRuns      int                `json:"totalRuns"`
	TotalProcessed int                `json:"totalProcessed"`
	TotalSucceeded int                `json:"totalSucceeded"`
	TotalFailed    int                `json:"totalFailed"`
	TotalTagged    int                `json:"totalTagged"`
}

func NewLinkCrawlerManager(dbPath string, interval time.Duration, classifyTags bool) *LinkCrawlerManager {
	if interval <= 0 {
		interval = time.Minute
	}
	return &LinkCrawlerManager{
		interval:     interval,
		dbPath:       dbPath,
		classifyTags: classifyTags,
	}
}

func (m *LinkCrawlerManager) Start(parent context.Context) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return false
	}
	ctx, cancel := context.WithCancel(parent)
	m.running = true
	m.cancel = cancel
	go m.runLoop(ctx)
	return true
}

func (m *LinkCrawlerManager) Stop() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return false
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	return true
}

func (m *LinkCrawlerManager) RunOnce(ctx context.Context) (*LinkCrawlerReport, error) {
	opts := LinkCrawlerOptions{Limit: 5}
	if m.classifyTags {
		opts.Classifier = DefaultLinkAnalysisClassifier
	}

	m.mu.Lock()
	m.lastStartedAt = time.Now().UTC()
	m.lastError = ""
	m.mu.Unlock()

	report, err := CrawlPendingLinks(ctx, m.dbPath, opts)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastFinishedAt = time.Now().UTC()
	if err != nil {
		m.lastError = err.Error()
		return nil, err
	}
	m.lastReport = report
	m.totalRuns++
	m.totalProcessed += report.Processed
	m.totalSucceeded += report.Succeeded
	m.totalFailed += report.Failed
	m.totalTagged += report.Tagged
	if len(report.Errors) > 0 {
		m.lastError = report.Errors[len(report.Errors)-1]
		return report, nil
	}
	m.lastError = ""
	return report, nil
}

func (m *LinkCrawlerManager) Status() LinkCrawlerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	status := LinkCrawlerStatus{
		Running:        m.running,
		IntervalMs:     m.interval.Milliseconds(),
		ClassifyTags:   m.classifyTags,
		DBPath:         m.dbPath,
		LastError:      m.lastError,
		LastReport:     cloneCrawlerReport(m.lastReport),
		TotalRuns:      m.totalRuns,
		TotalProcessed: m.totalProcessed,
		TotalSucceeded: m.totalSucceeded,
		TotalFailed:    m.totalFailed,
		TotalTagged:    m.totalTagged,
	}
	if !m.lastStartedAt.IsZero() {
		status.LastStartedAt = m.lastStartedAt.Format(time.RFC3339)
	}
	if !m.lastFinishedAt.IsZero() {
		status.LastFinishedAt = m.lastFinishedAt.Format(time.RFC3339)
	}
	return status
}

func (m *LinkCrawlerManager) runLoop(ctx context.Context) {
	_, _ = m.RunOnce(ctx)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = m.RunOnce(ctx)
		}
	}
}

func cloneCrawlerReport(report *LinkCrawlerReport) *LinkCrawlerReport {
	if report == nil {
		return nil
	}
	clone := *report
	if report.Errors != nil {
		clone.Errors = append([]string(nil), report.Errors...)
	}
	return &clone
}
