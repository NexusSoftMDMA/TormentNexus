package sales

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// OrderDB defines the database interface needed by the OrderProcessor.
type OrderDB interface {
	GetCompanyByID(ctx context.Context, id int64) (*db.Company, error)
}

// Processor coordinates the transition from a won deal to an active order.
type Processor struct {
	db	OrderDB
	billing	billing.BillingClient
	crm	crm.CRMClient
}

// NewOrderProcessor creates a new Processor instance.
func NewOrderProcessor(database OrderDB, billingClient billing.BillingClient, crmClient crm.CRMClient) *Processor {
	return &Processor{
		db:		database,
		billing:	billingClient,
		crm:		crmClient,
	}
}

// ProcessOrder handles fulfillment for a deal that has been closed won.
func (p *Processor) ProcessOrder(ctx context.Context, deal db.Deal) error {
	slog.Info(fmt.Sprintf("OrderProcessor: Processing fulfillment for deal %d", deal.ID))

	company, err := p.db.GetCompanyByID(ctx, deal.CompanyID)
	if err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}

	// 1. Generate Invoice
	invoiceID, err := p.billing.CreateInvoice(ctx, deal, *company)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	slog.Info(fmt.Sprintf("OrderProcessor: Invoice %s created for deal %d", invoiceID, deal.ID))

	// 2. Synchronize with CRM
	err = p.crm.SyncInteraction(ctx, deal.ID, fmt.Sprintf("Order processed. Invoice: %s", invoiceID))
	if err != nil {
		slog.Info(fmt.Sprintf("OrderProcessor Warning: CRM sync failed: %v", err))
	}

	// 3. (Optional) Trigger Provisioning
	// In a full implementation, this might call a provisioning service.
	slog.Info(fmt.Sprintf("OrderProcessor: Fulfillment complete for deal %d", deal.ID))

	return nil
}
