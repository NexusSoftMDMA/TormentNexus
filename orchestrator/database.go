package orchestrator

import (
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ---------------------------------------------------------
// TormentNexusa ORM Native Structural Parity (TormentNexus Go Daemon)
// ---------------------------------------------------------

// KeeperLog replaces TS KeeperLog model natively
type KeeperLog struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	SessionId string    `gorm:"type:varchar(255)" json:"sessionId"`
	Type      string    `gorm:"type:varchar(255)" json:"type"`
	Message   string    `gorm:"type:text" json:"message"`
	Metadata  string    `gorm:"type:text" json:"metadata,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// KeeperSettings replaces TS KeeperSettings model natively
type KeeperSettings struct {
	ID                         string    `gorm:"primaryKey;type:varchar(255);default:'default'" json:"id"`
	IsEnabled                  bool      `gorm:"default:false" json:"isEnabled"`
	AutoSwitch                 bool      `gorm:"default:false" json:"autoSwitch"`
	CheckIntervalSeconds       int       `gorm:"default:60" json:"checkIntervalSeconds"`
	InactivityThresholdMinutes int       `gorm:"default:10" json:"inactivityThresholdMinutes"`
	ActiveWorkThresholdMinutes int       `gorm:"default:5" json:"activeWorkThresholdMinutes"`
	Messages                   string    `gorm:"type:text" json:"messages"`
	CustomMessages             string    `gorm:"type:text" json:"customMessages"`
	SmartPilotEnabled          bool      `gorm:"default:false" json:"smartPilotEnabled"`
	SupervisorProvider         string    `gorm:"type:varchar(255);default:'openai'" json:"supervisorProvider"`
	SupervisorApiKey           string    `gorm:"type:text" json:"supervisorApiKey,omitempty"`
	JulesApiKey                string    `gorm:"type:text" json:"julesApiKey,omitempty"`
	SupervisorModel            string    `gorm:"type:varchar(255);default:'gpt-4o'" json:"supervisorModel"`
	ContextMessageCount        int       `gorm:"default:10" json:"contextMessageCount"`
	ResumePaused               bool      `gorm:"default:false" json:"resumePaused"`
	UserId                     string    `gorm:"type:varchar(255)" json:"userId,omitempty"`
	UpdatedAt                  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// Session replaces TS Session model natively
type Session struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	SessionToken string    `gorm:"unique;type:varchar(255)" json:"sessionToken"`
	UserId       string    `gorm:"type:varchar(255)" json:"userId"`
	Expires      time.Time `json:"expires"`
	// Additional state mock columns utilized natively by Dashboard
	Title     string    `gorm:"type:varchar(255);default:'Native Daemon Session'" json:"title"`
	Status    string    `gorm:"type:varchar(255);default:'active'" json:"status"`
	RawState  string    `gorm:"type:varchar(255);default:'ACTIVE'" json:"rawState"`
	SourceId  string    `gorm:"type:varchar(255);default:'google/jules'" json:"sourceId"`
	Branch    string    `gorm:"type:varchar(255);default:'main'" json:"branch"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// QueueJob replaces TS QueueJob model natively
type QueueJob struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Type        string    `gorm:"type:varchar(255)" json:"type"`
	Payload     string    `gorm:"type:text" json:"payload"`
	Status      string    `gorm:"index;type:varchar(255);default:'pending'" json:"status"`
	Attempts    int       `gorm:"default:0" json:"attempts"`
	MaxAttempts int       `gorm:"default:3" json:"maxAttempts"`
	LastError   string    `gorm:"type:text" json:"lastError,omitempty"`
	RunAt       time.Time `gorm:"index;autoCreateTime" json:"runAt"` // SQLite pure Go fallback
	StartedAt   time.Time `json:"startedAt,omitempty"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// MemoryChunk tracks relational vector text references
type MemoryChunk struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	SessionId string    `gorm:"index;type:varchar(255)" json:"sessionId"`
	Type      string    `gorm:"type:varchar(255)" json:"type"`
	Content   string    `gorm:"type:text" json:"content"`
	Metadata  string    `gorm:"type:text" json:"metadata,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// CodeChunk tracks the literal RAG repository vectors
type CodeChunk struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	WorkspaceId string    `gorm:"index;type:varchar(255)" json:"workspaceId"`
	Filepath    string    `gorm:"type:text" json:"filepath"`
	Content     string    `gorm:"type:text" json:"content"`
	StartLine   int       `json:"startLine"`
	EndLine     int       `json:"endLine"`
	Checksum    string    `gorm:"type:varchar(255)" json:"checksum"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// InitDatabase constructs the CGO-Free pure Go SQLite connection linking TormentNexusa Parity!
func InitDatabase(dbPath string) error {
	log.Printf("[Database] Constructing Native ORM Mapping against %s", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("ORM spinup failure on %s: %w", dbPath, err)
	}

	DB = db

	// Auto-Migrate structurally maps our pure Go structs backwards compatibly atop TormentNexusa tables!
	err = db.AutoMigrate(
		&KeeperLog{},
		&KeeperSettings{},
		&Session{},
		&QueueJob{},
		&MemoryChunk{},
		&CodeChunk{},
	)
	if err != nil {
		return fmt.Errorf("SQLite Auto-Migration rejection: %w", err)
	}

	log.Printf("[Database] ORM AutoMigration Sync Validated. TormentNexusa Legacy completely deprecated.")
	return nil
}
