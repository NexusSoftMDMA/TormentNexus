package supervisor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
	worktreegit "github.com/tormentnexushq/tormentnexus-go/internal/git"
)

type SessionState string

const (
	StateCreated    SessionState = "created"
	StateStarting   SessionState = "starting"
	StateRunning    SessionState = "running"
	StateStopping   SessionState = "stopping"
	StateStopped    SessionState = "stopped"
	StateFailed     SessionState = "error"
	StateRestarting SessionState = "restarting"
)

const (
	defaultRestartDelay  = 2 * time.Second
	defaultMaxLogEntries = 200
)

type SessionLogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Stream    string `json:"stream"`
	Message   string `json:"message"`
}

type SessionHealth struct {
	Status              string  `json:"status"`
	LastCheck           int64   `json:"lastCheck"`
	ConsecutiveFailures int     `json:"consecutiveFailures"`
	RestartCount        int     `json:"restartCount"`
	LastRestartAt       *int64  `json:"lastRestartAt,omitempty"`
	NextRestartAt       *int64  `json:"nextRestartAt,omitempty"`
	LastExitCode        *int    `json:"lastExitCode,omitempty"`
	LastExitSignal      *string `json:"lastExitSignal,omitempty"`
	ErrorMessage        *string `json:"errorMessage,omitempty"`
}

type SessionAttachInfo struct {
	ID                    string   `json:"id"`
	PID                   int      `json:"pid,omitempty"`
	Command               string   `json:"command"`
	Args                  []string `json:"args"`
	CWD                   string   `json:"cwd"`
	Status                string   `json:"status"`
	Attachable            bool     `json:"attachable"`
	AttachReadiness       string   `json:"attachReadiness"`
	AttachReadinessReason string   `json:"attachReadinessReason"`
}

type ExecutionPolicy struct {
	RequestedProfile   string  `json:"requestedProfile"`
	EffectiveProfile   string  `json:"effectiveProfile"`
	ShellID            *string `json:"shellId,omitempty"`
	ShellLabel         *string `json:"shellLabel,omitempty"`
	ShellFamily        *string `json:"shellFamily,omitempty"`
	ShellPath          *string `json:"shellPath,omitempty"`
	SupportsPowerShell bool    `json:"supportsPowerShell"`
	SupportsPosixShell bool    `json:"supportsPosixShell"`
	Reason             string  `json:"reason"`
}

type CreateSessionOptions struct {
	ID                  string
	Name                string
	CliType             string
	Command             string
	Args                []string
	Env                 map[string]string
	RequestedWorkingDir string
	WorkingDirectory    string
	ExecutionProfile    string
	AutoRestart         bool
	IsolateWorktree     bool
	Metadata            map[string]any
	MaxRestarts         int
}

type ManagerOptions struct {
	PersistencePath   string
	MaxPersisted      int
	MaxLogEntries     int
	AutoResumeOnStart bool
	RestartDelay      time.Duration
	WorktreeRoot      string
}

type RestoreStatus struct {
	LastRestoreAt        *int64 `json:"lastRestoreAt,omitempty"`
	RestoredSessionCount int    `json:"restoredSessionCount"`
	AutoResumeCount      int    `json:"autoResumeCount"`
}

type persistedState struct {
	Sessions []SupervisedSession `json:"sessions"`
	SavedAt  int64               `json:"savedAt"`
}

type SupervisedSession struct {
	ID                        string            `json:"id"`
	Name                      string            `json:"name"`
	CliType                   string            `json:"cliType"`
	Command                   string            `json:"command"`
	Args                      []string          `json:"args"`
	Env                       map[string]string `json:"env"`
	ExecutionProfile          string            `json:"executionProfile"`
	ExecutionPolicy           *ExecutionPolicy  `json:"executionPolicy,omitempty"`
	RequestedWorkingDirectory string            `json:"requestedWorkingDirectory"`
	WorkingDirectory          string            `json:"workingDirectory"`
	WorktreePath              string            `json:"worktreePath,omitempty"`
	AutoRestart               bool              `json:"autoRestart"`
	IsolateWorktree           bool              `json:"isolateWorktree"`
	State                     SessionState      `json:"status"`
	PID                       int               `json:"pid,omitempty"`
	RestartCount              int               `json:"restartCount"`
	MaxRestarts               int               `json:"maxRestartAttempts"`
	CreatedAt                 int64             `json:"createdAt"`
	StartedAt                 int64             `json:"startedAt,omitempty"`
	StoppedAt                 int64             `json:"stoppedAt,omitempty"`
	ScheduledRestartAt        int64             `json:"scheduledRestartAt,omitempty"`
	LastActivityAt            int64             `json:"lastActivityAt"`
	LastError                 string            `json:"lastError,omitempty"`
	LastExitCode              int               `json:"lastExitCode,omitempty"`
	LastExitSignal            string            `json:"lastExitSignal,omitempty"`
	Metadata                  map[string]any    `json:"metadata"`
	Logs                      []SessionLogEntry `json:"logs"`

	health           SessionHealth   `json:"-"`
	cmd              *exec.Cmd       `json:"-"`
	manualStop       bool            `json:"-"`
	restartAfterStop bool            `json:"-"`
	restartTimer     *time.Timer     `json:"-"`
	restartContext   context.Context `json:"-"`
}

type Manager struct {
	sessions          map[string]*SupervisedSession
	mu                sync.RWMutex
	persistencePath   string
	maxPersisted      int
	maxLogEntries     int
	autoResumeOnStart bool
	restartDelay      time.Duration
	restoreStatus     RestoreStatus
	worktreeManager   *worktreegit.WorktreeManager
	monitor           *ConversationMonitor
}

func NewManager(options ...ManagerOptions) *Manager {
	cfg := ManagerOptions{}
	if len(options) > 0 {
		cfg = options[0]
	}
	if cfg.MaxPersisted <= 0 {
		cfg.MaxPersisted = 100
	}
	if cfg.MaxLogEntries <= 0 {
		cfg.MaxLogEntries = defaultMaxLogEntries
	}
	if cfg.RestartDelay <= 0 {
		cfg.RestartDelay = defaultRestartDelay
	}
	manager := &Manager{
		sessions:          make(map[string]*SupervisedSession),
		persistencePath:   strings.TrimSpace(cfg.PersistencePath),
		maxPersisted:      cfg.MaxPersisted,
		maxLogEntries:     cfg.MaxLogEntries,
		autoResumeOnStart: cfg.AutoResumeOnStart,
		restartDelay:      cfg.RestartDelay,
		restoreStatus: RestoreStatus{
			RestoredSessionCount: 0,
			AutoResumeCount:      0,
		},
	}
	if strings.TrimSpace(cfg.WorktreeRoot) != "" {
		manager.worktreeManager = worktreegit.NewWorktreeManager(cfg.WorktreeRoot)
	}
	manager.restoreSessions()
	return manager
}

func (m *Manager) SetPredictor(predictor *mcp.ToolPredictor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.monitor = NewConversationMonitor(m, predictor)
	m.monitor.Start(context.Background())
}

func (m *Manager) CreateSession(id, command string, args []string, env map[string]string, cwd string, maxRestarts int) (*SupervisedSession, error) {
	return m.CreateSessionWithOptions(CreateSessionOptions{
		ID:                  id,
		Name:                id,
		CliType:             command,
		Command:             command,
		Args:                args,
		Env:                 env,
		RequestedWorkingDir: cwd,
		WorkingDirectory:    cwd,
		ExecutionProfile:    "auto",
		AutoRestart:         true,
		IsolateWorktree:     false,
		Metadata:            map[string]any{},
		MaxRestarts:         maxRestarts,
	})
}

func (m *Manager) CreateSessionWithOptions(input CreateSessionOptions) (*SupervisedSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("session-%d", time.Now().UTC().UnixNano())
	}
	if _, exists := m.sessions[id]; exists {
		return nil, fmt.Errorf("session %s already exists", id)
	}

	command := strings.TrimSpace(input.Command)
	if command == "" {
		return nil, fmt.Errorf("session %s has no command", id)
	}
	cliType := strings.TrimSpace(input.CliType)
	if cliType == "" {
		cliType = command
	}
	workingDirectory := strings.TrimSpace(input.WorkingDirectory)
	if workingDirectory == "" {
		workingDirectory = strings.TrimSpace(input.RequestedWorkingDir)
	}
	if workingDirectory == "" {
		workingDirectory = "."
	}
	requestedWorkingDirectory := strings.TrimSpace(input.RequestedWorkingDir)
	if requestedWorkingDirectory == "" {
		requestedWorkingDirectory = workingDirectory
	}
	usesWorktree, worktreePath, worktreeReason := m.allocateWorktreeLocked(id, requestedWorkingDirectory, input.IsolateWorktree)
	if usesWorktree {
		workingDirectory = worktreePath
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = fmt.Sprintf("%s-%s", cliType, shortenID(id))
	}
	maxRestarts := input.MaxRestarts
	if maxRestarts < 0 {
		maxRestarts = 0
	}
	executionProfile := strings.TrimSpace(input.ExecutionProfile)
	if executionProfile == "" {
		executionProfile = "auto"
	}
	autoRestart := input.AutoRestart
	if !input.AutoRestart && input.MaxRestarts == 0 {
		autoRestart = false
	} else if input.AutoRestart || input.MaxRestarts > 0 {
		autoRestart = true
	}
	executionPolicy := detectExecutionPolicy(executionProfile)
	mergedEnv := cloneEnv(input.Env)
	for key, value := range buildExecutionPolicyEnv(executionPolicy) {
		mergedEnv[key] = value
	}
	now := nowMillis()
	session := &SupervisedSession{
		ID:                        id,
		Name:                      name,
		CliType:                   cliType,
		Command:                   command,
		Args:                      append([]string(nil), input.Args...),
		Env:                       mergedEnv,
		ExecutionProfile:          executionProfile,
		ExecutionPolicy:           executionPolicy,
		RequestedWorkingDirectory: requestedWorkingDirectory,
		WorkingDirectory:          workingDirectory,
		WorktreePath:              worktreePath,
		AutoRestart:               autoRestart,
		IsolateWorktree:           usesWorktree,
		State:                     StateCreated,
		RestartCount:              0,
		MaxRestarts:               maxRestarts,
		CreatedAt:                 now,
		LastActivityAt:            now,
		Metadata:                  cloneMetadata(input.Metadata),
		Logs:                      []SessionLogEntry{},
		health: SessionHealth{
			Status:              "degraded",
			LastCheck:           now,
			ConsecutiveFailures: 0,
			RestartCount:        0,
		},
	}
	m.sessions[id] = session
	m.appendLogLocked(session, "system", fmt.Sprintf("Session created for %s in %s", session.CliType, session.WorkingDirectory))
	if strings.TrimSpace(worktreeReason) != "" {
		m.appendLogLocked(session, "system", worktreeReason)
	}
	if executionPolicy != nil && strings.TrimSpace(executionPolicy.Reason) != "" {
		m.appendLogLocked(session, "system", fmt.Sprintf("Execution policy %s selected%s (%s)", executionPolicy.EffectiveProfile, executionPolicyLogShellSuffix(executionPolicy), executionPolicy.Reason))
	}
	clone := m.cloneSessionLocked(session)
	m.persistLocked()
	return clone, nil
}

func (m *Manager) GetSession(id string) (*SupervisedSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[id]
	if !ok {
		return nil, false
	}
	return m.cloneSessionLocked(session), true
}

func (m *Manager) StartSession(ctx context.Context, id string) error {
	m.mu.Lock()
	session, exists := m.sessions[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("session %s not found", id)
	}
	if session.State == StateRunning || session.State == StateStarting {
		m.mu.Unlock()
		return nil
	}
	if session.restartTimer != nil {
		session.restartTimer.Stop()
		session.restartTimer = nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	session.restartContext = ctx
	session.manualStop = false
	session.restartAfterStop = false
	session.State = StateStarting
	session.ScheduledRestartAt = 0
	session.LastError = ""
	session.health.Status = "degraded"
	session.health.LastCheck = nowMillis()
	session.health.NextRestartAt = nil
	session.LastActivityAt = nowMillis()
	m.appendLogLocked(session, "system", fmt.Sprintf("Starting %s %s", session.Command, strings.Join(session.Args, " ")))
	m.persistLocked()
	clone := m.cloneSessionLocked(session)
	m.mu.Unlock()
	return m.runSession(ctx, clone)
}

func (m *Manager) runSession(ctx context.Context, session *SupervisedSession) error {
	cmd := exec.CommandContext(ctx, session.Command, session.Args...)
	cmd.Dir = session.WorkingDirectory
	cmd.Env = append([]string{}, os.Environ()...)
	for key, value := range session.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.markStartFailure(session.ID, fmt.Errorf("stdout pipe: %w", err))
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.markStartFailure(session.ID, fmt.Errorf("stderr pipe: %w", err))
		return err
	}
	if err := cmd.Start(); err != nil {
		m.markStartFailure(session.ID, err)
		return err
	}

	m.mu.Lock()
	live, ok := m.sessions[session.ID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session %s disappeared", session.ID)
	}
	live.cmd = cmd
	live.PID = cmd.Process.Pid
	live.State = StateRunning
	live.StartedAt = nowMillis()
	live.StoppedAt = 0
	live.LastActivityAt = nowMillis()
	live.health.Status = "healthy"
	live.health.LastCheck = nowMillis()
	live.health.ConsecutiveFailures = 0
	live.health.ErrorMessage = nil
	m.appendLogLocked(live, "system", fmt.Sprintf("Spawned process %d", live.PID))
	m.persistLocked()
	m.mu.Unlock()

	go m.streamOutput(session.ID, "stdout", stdout)
	go m.streamOutput(session.ID, "stderr", stderr)
	go m.waitForExit(ctx, session.ID, cmd)
	return nil
}

func (m *Manager) StopSession(id string) error {
	return m.StopSessionWithOptions(id, false)
}

func (m *Manager) StopSessionWithOptions(id string, force bool) error {
	m.mu.Lock()
	session, exists := m.sessions[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("session %s not found", id)
	}
	session.manualStop = true
	session.restartAfterStop = false
	if session.restartTimer != nil {
		session.restartTimer.Stop()
		session.restartTimer = nil
	}
	session.ScheduledRestartAt = 0
	session.health.NextRestartAt = nil
	session.health.Status = "degraded"
	session.health.LastCheck = nowMillis()
	if session.cmd == nil || session.cmd.Process == nil {
		session.State = StateStopped
		session.StoppedAt = nowMillis()
		m.appendLogLocked(session, "system", "Stop requested while no process was running.")
		m.persistLocked()
		m.mu.Unlock()
		return nil
	}
	session.State = StateStopping
	if force {
		m.appendLogLocked(session, "system", "Stopping process forcefully.")
	} else {
		m.appendLogLocked(session, "system", "Stopping process.")
	}
	process := session.cmd.Process
	m.persistLocked()
	m.mu.Unlock()
	if force {
		return process.Kill()
	}
	return process.Signal(os.Interrupt)
}

func (m *Manager) RestartSession(ctx context.Context, id string) error {
	m.mu.Lock()
	session, exists := m.sessions[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("session %s not found", id)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	session.restartContext = ctx
	if session.cmd == nil || session.cmd.Process == nil {
		session.State = StateRestarting
		m.appendLogLocked(session, "system", "Restart requested while idle; starting session.")
		m.persistLocked()
		m.mu.Unlock()
		return m.StartSession(ctx, id)
	}
	session.manualStop = true
	session.restartAfterStop = true
	session.State = StateRestarting
	m.appendLogLocked(session, "system", "Manual restart requested.")
	process := session.cmd.Process
	m.persistLocked()
	m.mu.Unlock()
	return process.Signal(os.Interrupt)
}

func (m *Manager) GetSessionLogs(id string, limit int) ([]SessionLogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	if limit <= 0 || limit > len(session.Logs) {
		limit = len(session.Logs)
	}
	result := make([]SessionLogEntry, 0, limit)
	for _, entry := range session.Logs[len(session.Logs)-limit:] {
		result = append(result, entry)
	}
	return result, nil
}

func (m *Manager) GetAttachInfo(id string) (*SessionAttachInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return buildAttachInfo(session), nil
}

func (m *Manager) GetSessionHealth(id string) (*SessionHealth, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	health := session.health
	return &health, nil
}

func (m *Manager) ListSessions() []SupervisedSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listSessionsLocked()
}

func (m *Manager) GetRestoreStatus() RestoreStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	status := m.restoreStatus
	return status
}

func (m *Manager) RestoreSessions() []SupervisedSession {
	m.restoreSessions()
	return m.ListSessions()
}

func (m *Manager) Shutdown() error {
	m.mu.Lock()
	for _, session := range m.sessions {
		if session.restartTimer != nil {
			session.restartTimer.Stop()
			session.restartTimer = nil
		}
		if session.cmd != nil && session.cmd.Process != nil {
			session.manualStop = true
			_ = session.cmd.Process.Signal(os.Interrupt)
		}
	}
	m.persistLocked()
	m.mu.Unlock()
	return nil
}

func (m *Manager) restoreSessions() {
	state := m.readPersistedState()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[string]*SupervisedSession)
	autoResumeIDs := make([]string, 0)
	for _, persisted := range state.Sessions {
		normalized, shouldAutoResume := m.normalizeRestoredSession(persisted)
		m.sessions[normalized.ID] = normalized
		if shouldAutoResume {
			autoResumeIDs = append(autoResumeIDs, normalized.ID)
		}
	}
	lastRestoreAt := nowMillis()
	m.restoreStatus = RestoreStatus{
		LastRestoreAt:        &lastRestoreAt,
		RestoredSessionCount: len(m.sessions),
		AutoResumeCount:      len(autoResumeIDs),
	}
	for _, sessionID := range autoResumeIDs {
		go func(id string) {
			_ = m.StartSession(context.Background(), id)
		}(sessionID)
	}
}

func (m *Manager) allocateWorktreeLocked(sessionID string, requestedWorkingDirectory string, requestedIsolation bool) (bool, string, string) {
	if !requestedIsolation {
		return false, "", ""
	}
	if m.worktreeManager == nil {
		return false, "", "Worktree isolation requested, but no Go worktree manager is configured; continuing without isolation."
	}
	if !m.shouldUseWorktreeLocked(requestedWorkingDirectory) {
		return false, "", ""
	}
	worktreePath, err := m.worktreeManager.CreateTaskEnvironment(sessionID)
	if err != nil {
		return false, "", "Worktree isolation requested, but native Go worktree creation failed; continuing without isolation: " + err.Error()
	}
	return true, worktreePath, "Worktree isolation enabled at " + worktreePath
}

func (m *Manager) shouldUseWorktreeLocked(workingDirectory string) bool {
	resolvedDir, err := filepath.Abs(strings.TrimSpace(workingDirectory))
	if err != nil {
		resolvedDir = filepath.Clean(strings.TrimSpace(workingDirectory))
	}
	if strings.TrimSpace(resolvedDir) == "" {
		return false
	}
	for _, session := range m.sessions {
		if !sessionUsesActiveWorkspace(session) {
			continue
		}
		activeDir, err := filepath.Abs(strings.TrimSpace(session.WorkingDirectory))
		if err != nil {
			activeDir = filepath.Clean(strings.TrimSpace(session.WorkingDirectory))
		}
		if activeDir == resolvedDir {
			return true
		}
	}
	return false
}

func sessionUsesActiveWorkspace(session *SupervisedSession) bool {
	switch session.State {
	case StateCreated, StateStarting, StateRunning, StateRestarting:
		return true
	default:
		return false
	}
}

func (m *Manager) normalizeRestoredSession(session SupervisedSession) (*SupervisedSession, bool) {
	status := session.State
	shouldAutoResume := false
	switch session.State {
	case StateRunning, StateStarting, StateStopping:
		if m.autoResumeOnStart {
			status = StateRestarting
			shouldAutoResume = true
		} else {
			status = StateStopped
		}
	case StateRestarting:
		if m.autoResumeOnStart {
			status = StateRestarting
			shouldAutoResume = true
		} else {
			status = StateStopped
		}
	}
	now := nowMillis()
	normalizedPolicy := normalizeExecutionPolicy(session.ExecutionPolicy)
	restored := &SupervisedSession{
		ID:                        strings.TrimSpace(session.ID),
		Name:                      strings.TrimSpace(session.Name),
		CliType:                   strings.TrimSpace(session.CliType),
		Command:                   strings.TrimSpace(session.Command),
		Args:                      append([]string(nil), session.Args...),
		Env:                       cloneEnv(session.Env),
		ExecutionProfile:          defaultString(session.ExecutionProfile, "auto"),
		ExecutionPolicy:           normalizedPolicy,
		RequestedWorkingDirectory: defaultString(session.RequestedWorkingDirectory, session.WorkingDirectory),
		WorkingDirectory:          defaultString(session.WorkingDirectory, session.RequestedWorkingDirectory),
		WorktreePath:              strings.TrimSpace(session.WorktreePath),
		AutoRestart:               session.AutoRestart,
		IsolateWorktree:           session.IsolateWorktree,
		State:                     status,
		PID:                       0,
		RestartCount:              maxInt(session.RestartCount, 0),
		MaxRestarts:               maxInt(session.MaxRestarts, 0),
		CreatedAt:                 defaultInt64(session.CreatedAt, now),
		StartedAt:                 session.StartedAt,
		StoppedAt:                 session.StoppedAt,
		ScheduledRestartAt:        0,
		LastActivityAt:            defaultInt64(session.LastActivityAt, now),
		LastError:                 strings.TrimSpace(session.LastError),
		LastExitCode:              session.LastExitCode,
		LastExitSignal:            strings.TrimSpace(session.LastExitSignal),
		Metadata:                  cloneMetadata(session.Metadata),
		Logs:                      append([]SessionLogEntry(nil), session.Logs...),
		health: SessionHealth{
			Status:              defaultRestoredHealthStatus(status),
			LastCheck:           now,
			ConsecutiveFailures: 0,
			RestartCount:        maxInt(session.RestartCount, 0),
			LastExitCode:        optionalIntPointer(session.LastExitCode),
			LastExitSignal:      stringPointer(session.LastExitSignal),
			ErrorMessage:        stringPointer(session.LastError),
		},
	}
	if restored.Name == "" {
		restored.Name = fmt.Sprintf("%s-%s", defaultString(restored.CliType, restored.Command), shortenID(restored.ID))
	}
	if restored.CliType == "" {
		restored.CliType = restored.Command
	}
	if restored.WorkingDirectory == "" {
		restored.WorkingDirectory = "."
	}
	if restored.RequestedWorkingDirectory == "" {
		restored.RequestedWorkingDirectory = restored.WorkingDirectory
	}
	if len(restored.Logs) > m.maxLogEntries {
		restored.Logs = append([]SessionLogEntry(nil), restored.Logs[len(restored.Logs)-m.maxLogEntries:]...)
	}
	if restored.State == StateFailed {
		restored.health.ConsecutiveFailures = 1
	}
	return restored, shouldAutoResume
}

func (m *Manager) markStartFailure(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return
	}
	message := strings.TrimSpace(err.Error())
	session.cmd = nil
	session.PID = 0
	session.State = StateFailed
	session.LastError = message
	session.LastActivityAt = nowMillis()
	session.health.Status = "crashed"
	session.health.LastCheck = nowMillis()
	session.health.ConsecutiveFailures++
	session.health.ErrorMessage = stringPointer(message)
	m.appendLogLocked(session, "system", "Start failed: "+message)
	m.persistLocked()
}

func (m *Manager) streamOutput(id, stream string, reader interface{ Read([]byte) (int, error) }) {
	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		m.recordOutput(id, stream, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		m.recordOutput(id, "system", fmt.Sprintf("%s stream error: %v", stream, err))
	}
}

func (m *Manager) recordOutput(id, stream, message string) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(message, "\r", ""))
	if trimmed == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return
	}
	session.LastActivityAt = nowMillis()
	session.health.LastCheck = nowMillis()
	session.health.Status = "healthy"
	m.appendLogLocked(session, stream, trimmed)
	m.persistLocked()
}

func (m *Manager) waitForExit(ctx context.Context, id string, cmd *exec.Cmd) {
	err := cmd.Wait()
	m.mu.Lock()
	session, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	session.cmd = nil
	session.PID = 0
	session.LastActivityAt = nowMillis()
	session.StoppedAt = nowMillis()
	session.health.LastCheck = nowMillis()
	codePtr, signalPtr := exitDetails(err)
	if codePtr != nil {
		session.LastExitCode = *codePtr
		session.health.LastExitCode = codePtr
	}
	if signalPtr != nil {
		session.LastExitSignal = *signalPtr
		session.health.LastExitSignal = signalPtr
	}
	if session.restartAfterStop {
		session.restartAfterStop = false
		session.manualStop = false
		m.scheduleRestartLocked(session, ctx, "manual restart requested")
		m.persistLocked()
		m.mu.Unlock()
		return
	}
	if session.manualStop {
		session.manualStop = false
		session.State = StateStopped
		session.ScheduledRestartAt = 0
		session.health.Status = "degraded"
		session.health.NextRestartAt = nil
		session.health.ErrorMessage = nil
		m.appendLogLocked(session, "system", "Process stopped.")
		m.persistLocked()
		m.mu.Unlock()
		return
	}
	if err != nil {
		message := strings.TrimSpace(err.Error())
		session.LastError = message
		session.health.ConsecutiveFailures++
		session.health.ErrorMessage = stringPointer(message)
		if session.AutoRestart && session.RestartCount < session.MaxRestarts {
			session.RestartCount++
			session.health.RestartCount = session.RestartCount
			m.scheduleRestartLocked(session, ctx, message)
			m.persistLocked()
			m.mu.Unlock()
			return
		}
		session.State = StateFailed
		session.ScheduledRestartAt = 0
		session.health.NextRestartAt = nil
		session.health.Status = "crashed"
		m.appendLogLocked(session, "system", "Process exited with error: "+message)
		m.persistLocked()
		m.mu.Unlock()
		return
	}
	session.State = StateStopped
	session.ScheduledRestartAt = 0
	session.LastError = ""
	session.health.Status = "degraded"
	session.health.NextRestartAt = nil
	session.health.ErrorMessage = nil
	m.appendLogLocked(session, "system", "Process exited cleanly.")
	m.persistLocked()
	m.mu.Unlock()
}

func (m *Manager) scheduleRestartLocked(session *SupervisedSession, ctx context.Context, reason string) {
	restartAt := nowMillis() + m.restartDelay.Milliseconds()
	session.State = StateRestarting
	session.ScheduledRestartAt = restartAt
	session.health.Status = "degraded"
	session.health.LastCheck = nowMillis()
	session.health.RestartCount = session.RestartCount
	session.health.LastRestartAt = int64Pointer(nowMillis())
	session.health.NextRestartAt = int64Pointer(restartAt)
	m.appendLogLocked(session, "system", fmt.Sprintf("Restart scheduled: %s", reason))
	if session.restartTimer != nil {
		session.restartTimer.Stop()
	}
	restartCtx := ctx
	if restartCtx == nil {
		restartCtx = session.restartContext
	}
	if restartCtx == nil {
		restartCtx = context.Background()
	}
	session.restartContext = restartCtx
	session.restartTimer = time.AfterFunc(m.restartDelay, func() {
		_ = m.StartSession(restartCtx, session.ID)
	})
}

func (m *Manager) appendLogLocked(session *SupervisedSession, stream, message string) {
	entry := SessionLogEntry{
		Timestamp: nowMillis(),
		Stream:    stream,
		Message:   message,
	}
	session.Logs = append(session.Logs, entry)
	if len(session.Logs) > m.maxLogEntries {
		session.Logs = append([]SessionLogEntry(nil), session.Logs[len(session.Logs)-m.maxLogEntries:]...)
	}
	session.LastActivityAt = entry.Timestamp
}

func (m *Manager) listSessionsLocked() []SupervisedSession {
	list := make([]SupervisedSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		list = append(list, *m.cloneSessionLocked(session))
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].CreatedAt == list[j].CreatedAt {
			return list[i].ID < list[j].ID
		}
		return list[i].CreatedAt < list[j].CreatedAt
	})
	return list
}

func (m *Manager) cloneSessionLocked(session *SupervisedSession) *SupervisedSession {
	clone := *session
	clone.Args = append([]string(nil), session.Args...)
	clone.Env = cloneEnv(session.Env)
	clone.ExecutionPolicy = cloneExecutionPolicy(session.ExecutionPolicy)
	clone.Metadata = cloneMetadata(session.Metadata)
	clone.Logs = append([]SessionLogEntry(nil), session.Logs...)
	clone.cmd = nil
	clone.restartTimer = nil
	clone.restartContext = nil
	return &clone
}

func (m *Manager) readPersistedState() persistedState {
	if strings.TrimSpace(m.persistencePath) == "" {
		return persistedState{Sessions: []SupervisedSession{}, SavedAt: nowMillis()}
	}
	raw, err := os.ReadFile(m.persistencePath)
	if err != nil {
		return persistedState{Sessions: []SupervisedSession{}, SavedAt: nowMillis()}
	}
	var state persistedState
	if err := json.Unmarshal(raw, &state); err != nil {
		return persistedState{Sessions: []SupervisedSession{}, SavedAt: nowMillis()}
	}
	if state.Sessions == nil {
		state.Sessions = []SupervisedSession{}
	}
	if state.SavedAt == 0 {
		state.SavedAt = nowMillis()
	}
	return state
}

func (m *Manager) persistLocked() {
	if strings.TrimSpace(m.persistencePath) == "" {
		return
	}
	state := persistedState{
		Sessions: m.listSessionsLocked(),
		SavedAt:  nowMillis(),
	}
	if len(state.Sessions) > m.maxPersisted {
		state.Sessions = append([]SupervisedSession(nil), state.Sessions[len(state.Sessions)-m.maxPersisted:]...)
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(m.persistencePath), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(m.persistencePath, raw, 0o644)
}

func buildAttachInfo(session *SupervisedSession) *SessionAttachInfo {
	attachReadiness := "unavailable"
	attachReason := "error"
	hasPID := session.PID > 0
	switch session.State {
	case StateRunning:
		if hasPID {
			attachReadiness = "ready"
			attachReason = "running-with-pid"
		} else {
			attachReadiness = "unavailable"
			attachReason = "no-pid"
		}
	case StateStarting:
		attachReadiness = "pending"
		attachReason = "starting"
	case StateRestarting:
		attachReadiness = "pending"
		attachReason = "restarting"
	case StateStopping:
		attachReadiness = "pending"
		attachReason = "stopping"
	case StateStopped:
		attachReadiness = "unavailable"
		attachReason = "stopped"
	case StateCreated:
		attachReadiness = "unavailable"
		attachReason = "created"
	case StateFailed:
		attachReadiness = "unavailable"
		attachReason = "error"
	}
	return &SessionAttachInfo{
		ID:                    session.ID,
		PID:                   session.PID,
		Command:               session.Command,
		Args:                  append([]string(nil), session.Args...),
		CWD:                   session.WorkingDirectory,
		Status:                string(session.State),
		Attachable:            session.State == StateRunning && hasPID,
		AttachReadiness:       attachReadiness,
		AttachReadinessReason: attachReason,
	}
}

func cloneEnv(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneMetadata(source map[string]any) map[string]any {
	if len(source) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneExecutionPolicy(source *ExecutionPolicy) *ExecutionPolicy {
	if source == nil {
		return nil
	}
	clone := *source
	clone.ShellID = cloneStringPointer(source.ShellID)
	clone.ShellLabel = cloneStringPointer(source.ShellLabel)
	clone.ShellFamily = cloneStringPointer(source.ShellFamily)
	clone.ShellPath = cloneStringPointer(source.ShellPath)
	return &clone
}

func normalizeExecutionPolicy(source *ExecutionPolicy) *ExecutionPolicy {
	if source == nil {
		return detectExecutionPolicy("auto")
	}
	clone := cloneExecutionPolicy(source)
	if clone == nil {
		return detectExecutionPolicy("auto")
	}
	clone.RequestedProfile = defaultString(clone.RequestedProfile, "auto")
	clone.EffectiveProfile = defaultString(clone.EffectiveProfile, "fallback")
	clone.Reason = defaultString(clone.Reason, "Restored from Go supervisor persistence.")
	return clone
}

func detectExecutionPolicy(requestedProfile string) *ExecutionPolicy {
	requested := defaultString(requestedProfile, "auto")
	powerShell := firstAvailableShell([]shellCandidate{{ID: "pwsh", Label: "PowerShell 7", Family: "powershell", Command: "pwsh"}, {ID: "powershell", Label: "Windows PowerShell", Family: "powershell", Command: "powershell"}})
	posix := firstAvailableShell([]shellCandidate{{ID: "bash", Label: "Bash", Family: "posix", Command: "bash"}, {ID: "sh", Label: "POSIX sh", Family: "posix", Command: "sh"}, {ID: "wsl", Label: "Windows Subsystem for Linux", Family: "wsl", Command: "wsl"}})
	compatibility := firstAvailableShell([]shellCandidate{{ID: "cmd", Label: "Command Prompt", Family: "cmd", Command: envOrDefault("COMSPEC", "cmd")}, {ID: "powershell", Label: "Windows PowerShell", Family: "powershell", Command: "powershell"}})
	preferred := powerShell
	if preferred == nil {
		preferred = posix
	}
	if preferred == nil {
		preferred = compatibility
	}
	supportsPowerShell := powerShell != nil
	supportsPosixShell := posix != nil

	toPolicy := func(effective string, shell *shellCandidate, reason string) *ExecutionPolicy {
		return &ExecutionPolicy{
			RequestedProfile:   requested,
			EffectiveProfile:   effective,
			ShellID:            nullablePolicyString(shellField(shell, func(candidate *shellCandidate) string { return candidate.ID })),
			ShellLabel:         nullablePolicyString(shellField(shell, func(candidate *shellCandidate) string { return candidate.Label })),
			ShellFamily:        nullablePolicyString(shellField(shell, func(candidate *shellCandidate) string { return candidate.Family })),
			ShellPath:          nullablePolicyString(shellField(shell, func(candidate *shellCandidate) string { return candidate.Path })),
			SupportsPowerShell: supportsPowerShell,
			SupportsPosixShell: supportsPosixShell,
			Reason:             reason,
		}
	}

	switch requested {
	case "powershell":
		if powerShell != nil {
			return toPolicy("powershell", powerShell, powerShell.Label+" selected because the session explicitly requested a PowerShell-native execution profile.")
		}
		return toPolicy("fallback", preferred, fallbackReason("A PowerShell shell was requested, but none verified", preferred))
	case "posix":
		if posix != nil {
			return toPolicy("posix", posix, posix.Label+" selected because the session requested POSIX-style pipelines or Unix-first tooling.")
		}
		return toPolicy("fallback", preferred, fallbackReason("A POSIX shell was requested, but none verified", preferred))
	case "compatibility":
		if runtime.GOOS == "windows" && compatibility != nil && compatibility.Family == "cmd" {
			return toPolicy("compatibility", compatibility, compatibility.Label+" selected for the most conservative compatibility posture on this host.")
		}
		return toPolicy("fallback", preferred, fallbackReason("Compatibility mode was requested, but no conservative shell profile was verified", preferred))
	default:
		if runtime.GOOS == "windows" && powerShell != nil {
			return toPolicy("powershell", powerShell, powerShell.Label+" selected automatically as TormentNexus's preferred Windows execution shell for general harness supervision.")
		}
		if posix != nil {
			return toPolicy("posix", posix, posix.Label+" selected automatically because no verified PowerShell shell was preferred for this host.")
		}
		return toPolicy("fallback", preferred, fallbackReason("Auto execution profile could not verify a strongly preferred shell", preferred))
	}
}

type shellCandidate struct {
	ID      string
	Label   string
	Family  string
	Command string
	Path    string
}

func firstAvailableShell(candidates []shellCandidate) *shellCandidate {
	for _, candidate := range candidates {
		command := strings.TrimSpace(candidate.Command)
		if command == "" {
			continue
		}
		resolved, err := exec.LookPath(command)
		if err != nil {
			continue
		}
		candidate.Path = resolved
		copy := candidate
		return &copy
	}
	return nil
}

func buildExecutionPolicyEnv(policy *ExecutionPolicy) map[string]string {
	if policy == nil {
		return map[string]string{}
	}
	env := map[string]string{
		"TORMENTNEXUS_EXECUTION_PROFILE_REQUESTED": policy.RequestedProfile,
		"TORMENTNEXUS_EXECUTION_PROFILE_EFFECTIVE": policy.EffectiveProfile,
		"TORMENTNEXUS_EXECUTION_SHELL_ID":          derefPolicyString(policy.ShellID),
		"TORMENTNEXUS_EXECUTION_SHELL_LABEL":       derefPolicyString(policy.ShellLabel),
		"TORMENTNEXUS_EXECUTION_SHELL_FAMILY":      derefPolicyString(policy.ShellFamily),
		"TORMENTNEXUS_EXECUTION_SHELL_PATH":        derefPolicyString(policy.ShellPath),
		"TORMENTNEXUS_EXECUTION_POLICY_REASON":     policy.Reason,
		"TORMENTNEXUS_SUPPORTS_POWERSHELL":         boolEnvValue(policy.SupportsPowerShell),
		"TORMENTNEXUS_SUPPORTS_POSIX_SHELL":        boolEnvValue(policy.SupportsPosixShell),
	}
	if policy.ShellPath != nil && strings.TrimSpace(*policy.ShellPath) != "" {
		env["SHELL"] = strings.TrimSpace(*policy.ShellPath)
		env["npm_config_script_shell"] = strings.TrimSpace(*policy.ShellPath)
		if policy.ShellFamily != nil && strings.TrimSpace(*policy.ShellFamily) == "cmd" {
			env["COMSPEC"] = strings.TrimSpace(*policy.ShellPath)
		}
	}
	return env
}

func executionPolicyLogShellSuffix(policy *ExecutionPolicy) string {
	if policy == nil || policy.ShellLabel == nil || strings.TrimSpace(*policy.ShellLabel) == "" {
		return ""
	}
	return " using " + strings.TrimSpace(*policy.ShellLabel)
}

func fallbackReason(prefix string, preferred *shellCandidate) string {
	if preferred == nil {
		return prefix + "; TormentNexus could not verify any shell on this host."
	}
	return prefix + "; falling back to " + preferred.Label + "."
}

func shellField(shell *shellCandidate, getter func(*shellCandidate) string) string {
	if shell == nil {
		return ""
	}
	return getter(shell)
}

func nullablePolicyString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func derefPolicyString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := strings.TrimSpace(*value)
	return &copy
}

func boolEnvValue(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}
	return fallback
}

func nowMillis() int64 {
	return time.Now().UTC().UnixMilli()
}

func shortenID(id string) string {
	trimmed := strings.TrimSpace(id)
	if len(trimmed) <= 8 {
		return trimmed
	}
	return trimmed[:8]
}

func stringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func int64Pointer(value int64) *int64 {
	return &value
}

func optionalIntPointer(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func exitDetails(err error) (*int, *string) {
	if err == nil {
		zero := 0
		return &zero, nil
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		message := strings.TrimSpace(err.Error())
		if message == "" {
			return nil, nil
		}
		return nil, &message
	}
	code := exitErr.ExitCode()
	var codePtr *int
	if code >= 0 {
		codeCopy := code
		codePtr = &codeCopy
	}
	message := strings.TrimSpace(exitErr.Error())
	if message == "" {
		return codePtr, nil
	}
	return codePtr, &message
}

func defaultRestoredHealthStatus(state SessionState) string {
	if state == StateFailed {
		return "crashed"
	}
	if state == StateRunning {
		return "healthy"
	}
	return "degraded"
}

func defaultString(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return strings.TrimSpace(fallback)
	}
	return trimmed
}

func defaultInt64(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}

func maxInt(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
