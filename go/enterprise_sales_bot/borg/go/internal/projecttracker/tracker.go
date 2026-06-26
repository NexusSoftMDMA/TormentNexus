package projecttracker

import (
	"os"
	"path/filepath"
	"sync"
)

type Project struct {
	Name     string `json:"name"`
	RootPath string `json:"rootPath"`
	Type     string `json:"type"`
	Language string `json:"language"`
}

type Service struct {
	mu       sync.RWMutex
	projects map[string]*Project
	root     string
}

func NewService(root string) *Service {
	return &Service{
		projects: make(map[string]*Project),
		root:     root,
	}
}

func (s *Service) DetectProjects() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projPath := filepath.Join(s.root, entry.Name())
		proj := s.detectProject(entry.Name(), projPath)
		if proj != nil {
			s.projects[entry.Name()] = proj
		}
	}
	return nil
}

func (s *Service) detectProject(name, path string) *Project {
	goMod := filepath.Join(path, "go.mod")
	packageJSON := filepath.Join(path, "package.json")
	pyProject := filepath.Join(path, "pyproject.toml")
	cargo := filepath.Join(path, "Cargo.toml")

	switch {
	case fileExists(goMod):
		return &Project{Name: name, RootPath: path, Type: "module", Language: "Go"}
	case fileExists(packageJSON):
		return &Project{Name: name, RootPath: path, Type: "module", Language: "TypeScript/JavaScript"}
	case fileExists(pyProject):
		return &Project{Name: name, RootPath: path, Type: "module", Language: "Python"}
	case fileExists(cargo):
		return &Project{Name: name, RootPath: path, Type: "module", Language: "Rust"}
	}
	return nil
}

func (s *Service) List() []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Project, 0, len(s.projects))
	for _, p := range s.projects {
		result = append(result, p)
	}
	return result
}

func (s *Service) Get(name string) *Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projects[name]
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
