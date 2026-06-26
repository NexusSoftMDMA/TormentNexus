package compat

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Catalog stores exact model-facing tool contracts.
type Catalog struct {
	mu       sync.RWMutex
	bySource map[string][]ToolContract
	byName   map[string][]ToolContract
}

func NewCatalog() *Catalog {
	return &Catalog{
		bySource: map[string][]ToolContract{},
		byName:   map[string][]ToolContract{},
	}
}

func (c *Catalog) Register(contract ToolContract) error {
	if strings.TrimSpace(contract.Source) == "" {
		return fmt.Errorf("tool contract source is required")
	}
	if strings.TrimSpace(contract.Name) == "" {
		return fmt.Errorf("tool contract name is required")
	}

	clone := contract.Clone()
	clone.Source = strings.TrimSpace(clone.Source)
	clone.Name = strings.TrimSpace(clone.Name)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.bySource[clone.Source] = append(c.bySource[clone.Source], clone)
	c.byName[clone.Name] = append(c.byName[clone.Name], clone)
	return nil
}

func (c *Catalog) MustRegister(contract ToolContract) {
	if err := c.Register(contract); err != nil {
		panic(err)
	}
}

func (c *Catalog) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := 0
	for _, contracts := range c.bySource {
		total += len(contracts)
	}
	return total
}

func (c *Catalog) Sources() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sources := make([]string, 0, len(c.bySource))
	for source := range c.bySource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	return sources
}

func (c *Catalog) ContractsBySource(source string) []ToolContract {
	c.mu.RLock()
	defer c.mu.RUnlock()

	contracts := c.bySource[strings.TrimSpace(source)]
	out := make([]ToolContract, 0, len(contracts))
	for _, contract := range contracts {
		out = append(out, contract.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (c *Catalog) Lookup(name string) []ToolContract {
	c.mu.RLock()
	defer c.mu.RUnlock()

	contracts := c.byName[strings.TrimSpace(name)]
	out := make([]ToolContract, 0, len(contracts))
	for _, contract := range contracts {
		out = append(out, contract.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source == out[j].Source {
			return out[i].Name < out[j].Name
		}
		return out[i].Source < out[j].Source
	})
	return out
}
