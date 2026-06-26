package plugins

import (
	"fmt"
	"plugin"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
)

// Registry holds custom implementations of core interfaces.
type Registry struct {
	Sources       map[string]scraper.LeadSource
	Enrichers     map[string]enrichment.EnrichmentSource
	Classifiers   map[string]communication.IntentClassifier
	Responders    map[string]communication.ResponseGenerator
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		Sources:       make(map[string]scraper.LeadSource),
		Enrichers:     make(map[string]enrichment.EnrichmentSource),
		Classifiers:   make(map[string]communication.IntentClassifier),
		Responders:    make(map[string]communication.ResponseGenerator),
	}
}

// RegisterSource adds a LeadSource to the registry.
func (r *Registry) RegisterSource(name string, source scraper.LeadSource) {
	r.Sources[name] = source
}

// RegisterEnricher adds an EnrichmentSource.
func (r *Registry) RegisterEnricher(name string, enricher enrichment.EnrichmentSource) {
	r.Enrichers[name] = enricher
}

// RegisterClassifier adds an IntentClassifier.
func (r *Registry) RegisterClassifier(name string, classifier communication.IntentClassifier) {
	r.Classifiers[name] = classifier
}

// RegisterResponder adds a ResponseGenerator.
func (r *Registry) RegisterResponder(name string, responder communication.ResponseGenerator) {
	r.Responders[name] = responder
}

// LoadGoPlugin attempts to load a Go plugin (.so) from the given path.
// It looks for exported variables named "Source", "Enricher", "Classifier", or "Responder".
func (r *Registry) LoadGoPlugin(name string, path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("open plugin %s: %w", path, err)
	}

	loaded := false

	if sym, err := p.Lookup("Source"); err == nil {
		if s, ok := sym.(scraper.LeadSource); ok {
			r.RegisterSource(name, s)
			loaded = true
		}
	}

	if sym, err := p.Lookup("Enricher"); err == nil {
		if e, ok := sym.(enrichment.EnrichmentSource); ok {
			r.RegisterEnricher(name, e)
			loaded = true
		}
	}

	if sym, err := p.Lookup("Classifier"); err == nil {
		if c, ok := sym.(communication.IntentClassifier); ok {
			r.RegisterClassifier(name, c)
			loaded = true
		}
	}

	if sym, err := p.Lookup("Responder"); err == nil {
		if res, ok := sym.(communication.ResponseGenerator); ok {
			r.RegisterResponder(name, res)
			loaded = true
		}
	}

	if !loaded {
		return fmt.Errorf("plugin %s exported no recognized symbols", path)
	}

	return nil
}
