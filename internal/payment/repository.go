package payment

import (
	"context"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// Payment operations
	CreatePayment(ctx context.Context, payment *models.Payment) error
	FindPaymentByID(ctx context.Context, id string) (*models.Payment, error)
	FindPaymentsByCompany(ctx context.Context, companyID string, filters PaymentFilters) ([]*models.Payment, int64, error)
	UpdatePayment(ctx context.Context, payment *models.Payment) error
	
	// Invoice operations
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	FindInvoiceByID(ctx context.Context, id string) (*models.Invoice, error)
	FindInvoicesByCompany(ctx context.Context, companyID string, filters InvoiceFilters) ([]*models.Invoice, int64, error)
	FindInvoicesByStatus(ctx context.Context, companyID string, status string) ([]*models.Invoice, error)
	FindOverdueInvoices(ctx context.Context, companyID string) ([]*models.Invoice, error)
	UpdateInvoice(ctx context.Context, invoice *models.Invoice) error
	
	// Subscription operations
	CreateSubscription(ctx context.Context, subscription *models.Subscription) error
	FindSubscriptionByID(ctx context.Context, id string) (*models.Subscription, error)
	FindSubscriptionsByCompany(ctx context.Context, companyID string) ([]*models.Subscription, error)
	FindActiveSubscriptions(ctx context.Context, companyID string) ([]*models.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription *models.Subscription) error
	
	// Statistics
	GetPaymentStats(ctx context.Context, companyID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetInvoiceStats(ctx context.Context, companyID string) (map[string]interface{}, error)
	GetRevenueStats(ctx context.Context, companyID string, period string) (map[string]interface{}, error)
}

// paymentRepository implements PaymentRepository interface
type paymentRepository struct {
	db              *gorm.DB
	optimizedQueries *OptimizedPaymentQueries
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{
		db:              db,
		optimizedQueries: NewOptimizedPaymentQueries(db),
	}
}

// CreatePayment creates a new payment
func (r *paymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// FindPaymentByID finds a payment by ID
func (r *paymentRepository) FindPaymentByID(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).
		Preload("Invoice").
		First(&payment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// FindPaymentsByCompany finds payments by company
func (r *paymentRepository) FindPaymentsByCompany(ctx context.Context, companyID string, filters PaymentFilters) ([]*models.Payment, int64, error) {
	return r.optimizedQueries.GetPaymentsByCompanyIDOptimized(ctx, companyID, filters)
}

// UpdatePayment updates a payment
func (r *paymentRepository) UpdatePayment(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// CreateInvoice creates a new invoice
func (r *paymentRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// FindInvoiceByID finds an invoice by ID
func (r *paymentRepository) FindInvoiceByID(ctx context.Context, id string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).
		Preload("Payments").
		First(&invoice, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindInvoicesByCompany finds invoices by company
func (r *paymentRepository) FindInvoicesByCompany(ctx context.Context, companyID string, filters InvoiceFilters) ([]*models.Invoice, int64, error) {
	return r.optimizedQueries.GetInvoicesByCompanyIDOptimized(ctx, companyID, filters)
}

// FindInvoicesByStatus finds invoices by status
func (r *paymentRepository) FindInvoicesByStatus(ctx context.Context, companyID string, status string) ([]*models.Invoice, error) {
	return r.optimizedQueries.GetInvoicesByStatusOptimized(ctx, companyID, status)
}

// FindOverdueInvoices finds overdue invoices
func (r *paymentRepository) FindOverdueInvoices(ctx context.Context, companyID string) ([]*models.Invoice, error) {
	return r.optimizedQueries.GetOverdueInvoicesOptimized(ctx, companyID)
}

// UpdateInvoice updates an invoice
func (r *paymentRepository) UpdateInvoice(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}

// CreateSubscription creates a new subscription
func (r *paymentRepository) CreateSubscription(ctx context.Context, subscription *models.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

// FindSubscriptionByID finds a subscription by ID
func (r *paymentRepository) FindSubscriptionByID(ctx context.Context, id string) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := r.db.WithContext(ctx).First(&subscription, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

// FindSubscriptionsByCompany finds subscriptions by company
func (r *paymentRepository) FindSubscriptionsByCompany(ctx context.Context, companyID string) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	if err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// FindActiveSubscriptions finds active subscriptions
func (r *paymentRepository) FindActiveSubscriptions(ctx context.Context, companyID string) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, "active").
		Order("created_at DESC").
		Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// UpdateSubscription updates a subscription
func (r *paymentRepository) UpdateSubscription(ctx context.Context, subscription *models.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

// GetPaymentStats gets payment statistics
func (r *paymentRepository) GetPaymentStats(ctx context.Context, companyID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalPayments int64   `gorm:"column:total_payments"`
		TotalAmount   float64 `gorm:"column:total_amount"`
	}
	
	query := `
		SELECT 
			COUNT(*) as total_payments,
			COALESCE(SUM(amount), 0) as total_amount
		FROM payments
		WHERE company_id = ? AND payment_date BETWEEN ? AND ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID, startDate, endDate).Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_payments": stats.TotalPayments,
		"total_amount":   stats.TotalAmount,
	}, nil
}

// GetInvoiceStats gets invoice statistics
func (r *paymentRepository) GetInvoiceStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalInvoices   int64   `gorm:"column:total_invoices"`
		PendingInvoices int64   `gorm:"column:pending_invoices"`
		PaidInvoices    int64   `gorm:"column:paid_invoices"`
		OverdueInvoices int64   `gorm:"column:overdue_invoices"`
		TotalAmount     float64 `gorm:"column:total_amount"`
	}
	
	query := `
		SELECT 
			COUNT(*) as total_invoices,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_invoices,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as paid_invoices,
			COUNT(CASE WHEN status = 'overdue' THEN 1 END) as overdue_invoices,
			COALESCE(SUM(total_amount), 0) as total_amount
		FROM invoices
		WHERE company_id = ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_invoices":   stats.TotalInvoices,
		"pending_invoices": stats.PendingInvoices,
		"paid_invoices":    stats.PaidInvoices,
		"overdue_invoices": stats.OverdueInvoices,
		"total_amount":     stats.TotalAmount,
	}, nil
}

// GetRevenueStats gets revenue statistics
func (r *paymentRepository) GetRevenueStats(ctx context.Context, companyID string, period string) (map[string]interface{}, error) {
	// Implementation would calculate revenue by period
	return map[string]interface{}{}, nil
}

