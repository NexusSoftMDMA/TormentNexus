package services

import (
	"log"
	"sync"
)

// ServiceLocator acts as the central Dependency Injection container for the TormentNexus Go orchestrator.
// It matches the TS version's centralized instantiation of singletons.
type ServiceLocator struct {
	ConfigManager *ConfigService
	MemoryManager *MemoryService
	Metrics       *MetricsService

	// Future:
	// SandboxProvider *SandboxService
}

var (
	instance *ServiceLocator
	once     sync.Once
)

// GetLocator returns the global singleton instance of the ServiceLocator.
func GetLocator() *ServiceLocator {
	once.Do(func() {
		log.Println("[Core:Locator] Bootstrapping global DI container...")
		instance = &ServiceLocator{}
		// Initialize the base services immediately
		instance.ConfigManager = NewConfigService()
		instance.MemoryManager = NewMemoryService()
		instance.Metrics = NewMetricsService()
	})
	return instance
}

// Below are stubs for the core services to fulfill the interface parity.

type ConfigService struct{}

func NewConfigService() *ConfigService { return &ConfigService{} }

type MemoryService struct{}

func NewMemoryService() *MemoryService { return &MemoryService{} }

type MetricsService struct{}

func NewMetricsService() *MetricsService { return &MetricsService{} }
