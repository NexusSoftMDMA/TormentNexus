package researcher

import (
	"context"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Crawler defines an interface for extracting technical insights from web sources.
type Crawler interface {
	Crawl(ctx context.Context, target string) (string, error)
}

// DossierProcessor defines an interface for synthesizing research findings.
type DossierProcessor interface {
	Process(findings []string) (string, error)
}

// Researcher coordinates technical deep research for leads.
type Researcher struct {
	db        *db.DB
	crawlers  []Crawler
	processor DossierProcessor
	crmClient crm.CRMClient
}

// NewResearcher creates a new Researcher instance.
func NewResearcher(database *db.DB, crawlers []Crawler, processor DossierProcessor, crmClient crm.CRMClient) *Researcher {
	return &Researcher{
		db:        database,
		crawlers:  crawlers,
		processor: processor,
		crmClient: crmClient,
	}
}
