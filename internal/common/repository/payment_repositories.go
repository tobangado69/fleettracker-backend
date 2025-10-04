package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// InvoiceRepositoryImpl implements the InvoiceRepository interface
type InvoiceRepositoryImpl struct {
	*BaseRepository[models.Invoice]
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &InvoiceRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Invoice](db),
	}
}

// GetByCompany retrieves invoices by company ID with pagination
func (r *InvoiceRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices by company: %w", err)
	}
	
	return invoices, nil
}

// GetByInvoiceNumber retrieves an invoice by invoice number
func (r *InvoiceRepositoryImpl) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found with number: %s", invoiceNumber)
		}
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}
	return &invoice, nil
}

// GetByStatus retrieves invoices by status with pagination
func (r *InvoiceRepositoryImpl) GetByStatus(ctx context.Context, status string, pagination Pagination) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	query := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices by status: %w", err)
	}
	
	return invoices, nil
}

// GetOverdueInvoices retrieves all overdue invoices
func (r *InvoiceRepositoryImpl) GetOverdueInvoices(ctx context.Context) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Where("status != ? AND due_date < ?", "paid", now).Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	
	return invoices, nil
}

// GetByDueDateRange retrieves invoices by due date range with pagination
func (r *InvoiceRepositoryImpl) GetByDueDateRange(ctx context.Context, startDate, endDate time.Time, pagination Pagination) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	query := r.db.WithContext(ctx).Where("due_date BETWEEN ? AND ?", startDate, endDate).Order("due_date ASC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices by due date range: %w", err)
	}
	
	return invoices, nil
}

// UpdatePaymentStatus updates the payment status of an invoice
func (r *InvoiceRepositoryImpl) UpdatePaymentStatus(ctx context.Context, invoiceID string, status string, paidAmount float64) error {
	updates := map[string]interface{}{
		"status": status,
		"paid_amount": paidAmount,
		"balance_amount": gorm.Expr("total_amount - ?", paidAmount),
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ?", invoiceID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	
	return nil
}

// PaymentRepositoryImpl implements the PaymentRepository interface
type PaymentRepositoryImpl struct {
	*BaseRepository[models.Payment]
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &PaymentRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Payment](db),
	}
}

// GetByCompany retrieves payments by company ID with pagination
func (r *PaymentRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Payment, error) {
	var payments []*models.Payment
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by company: %w", err)
	}
	
	return payments, nil
}

// GetBySubscription retrieves payments by subscription ID with pagination
func (r *PaymentRepositoryImpl) GetBySubscription(ctx context.Context, subscriptionID string, pagination Pagination) ([]*models.Payment, error) {
	var payments []*models.Payment
	query := r.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by subscription: %w", err)
	}
	
	return payments, nil
}

// GetByStatus retrieves payments by status with pagination
func (r *PaymentRepositoryImpl) GetByStatus(ctx context.Context, status string, pagination Pagination) ([]*models.Payment, error) {
	var payments []*models.Payment
	query := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by status: %w", err)
	}
	
	return payments, nil
}

// GetByPaymentMethod retrieves payments by payment method with pagination
func (r *PaymentRepositoryImpl) GetByPaymentMethod(ctx context.Context, paymentMethod string, pagination Pagination) ([]*models.Payment, error) {
	var payments []*models.Payment
	query := r.db.WithContext(ctx).Where("payment_method = ?", paymentMethod).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by payment method: %w", err)
	}
	
	return payments, nil
}

// GetByDateRange retrieves payments by date range with pagination
func (r *PaymentRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate time.Time, pagination Pagination) ([]*models.Payment, error) {
	var payments []*models.Payment
	query := r.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", startDate, endDate).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by date range: %w", err)
	}
	
	return payments, nil
}

// GetByReferenceNumber retrieves a payment by reference number
func (r *PaymentRepositoryImpl) GetByReferenceNumber(ctx context.Context, referenceNumber string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).Where("reference_number = ?", referenceNumber).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found with reference number: %s", referenceNumber)
		}
		return nil, fmt.Errorf("failed to get payment by reference number: %w", err)
	}
	return &payment, nil
}

// SubscriptionRepositoryImpl implements the SubscriptionRepository interface
type SubscriptionRepositoryImpl struct {
	*BaseRepository[models.Subscription]
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &SubscriptionRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Subscription](db),
	}
}

// GetByCompany retrieves subscriptions by company ID with pagination
func (r *SubscriptionRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by company: %w", err)
	}
	
	return subscriptions, nil
}

// GetActiveSubscriptions retrieves all active subscriptions
func (r *SubscriptionRepositoryImpl) GetActiveSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Where("is_active = true AND status = ? AND end_date > ?", "active", now).Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	
	return subscriptions, nil
}

// GetExpiringSubscriptions retrieves subscriptions expiring within specified days
func (r *SubscriptionRepositoryImpl) GetExpiringSubscriptions(ctx context.Context, days int) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	expiryDate := time.Now().AddDate(0, 0, days)
	
	if err := r.db.WithContext(ctx).Where("is_active = true AND end_date BETWEEN ? AND ?", time.Now(), expiryDate).Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get expiring subscriptions: %w", err)
	}
	
	return subscriptions, nil
}

// GetByStatus retrieves subscriptions by status with pagination
func (r *SubscriptionRepositoryImpl) GetByStatus(ctx context.Context, status string, pagination Pagination) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	query := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by status: %w", err)
	}
	
	return subscriptions, nil
}

// GetByPlanType retrieves subscriptions by plan type with pagination
func (r *SubscriptionRepositoryImpl) GetByPlanType(ctx context.Context, planType string, pagination Pagination) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	query := r.db.WithContext(ctx).Where("plan_type = ?", planType).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by plan type: %w", err)
	}
	
	return subscriptions, nil
}

// UpdateStatus updates the status of a subscription
func (r *SubscriptionRepositoryImpl) UpdateStatus(ctx context.Context, subscriptionID string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", subscriptionID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}
	return nil
}

// GetCompanyActiveSubscription retrieves the active subscription for a company
func (r *SubscriptionRepositoryImpl) GetCompanyActiveSubscription(ctx context.Context, companyID string) (*models.Subscription, error) {
	var subscription models.Subscription
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Where("company_id = ? AND is_active = true AND status = ? AND end_date > ?", 
		companyID, "active", now).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active subscription found for company: %s", companyID)
		}
		return nil, fmt.Errorf("failed to get company active subscription: %w", err)
	}
	
	return &subscription, nil
}
