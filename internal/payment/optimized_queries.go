package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// PaymentFilters represents filters for listing payments
type PaymentFilters struct {
	Status        *string    `json:"status" form:"status"`
	PaymentMethod *string    `json:"payment_method" form:"payment_method"`
	InvoiceID     *string    `json:"invoice_id" form:"invoice_id"`
	StartDate     *time.Time `json:"start_date" form:"start_date"`
	EndDate       *time.Time `json:"end_date" form:"end_date"`
	MinAmount     *float64   `json:"min_amount" form:"min_amount"`
	MaxAmount     *float64   `json:"max_amount" form:"max_amount"`
	
	// Pagination
	Page      int    `json:"page" form:"page" validate:"min=1"`
	Limit     int    `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy    string `json:"sort_by" form:"sort_by" validate:"oneof=created_at updated_at payment_date amount"`
	SortOrder string `json:"sort_order" form:"sort_order" validate:"oneof=asc desc"`
}

// InvoiceFilters represents filters for listing invoices
type InvoiceFilters struct {
	Status       *string    `json:"status" form:"status"`
	StartDate    *time.Time `json:"start_date" form:"start_date"`
	EndDate      *time.Time `json:"end_date" form:"end_date"`
	DueDateStart *time.Time `json:"due_date_start" form:"due_date_start"`
	DueDateEnd   *time.Time `json:"due_date_end" form:"due_date_end"`
	MinAmount    *float64   `json:"min_amount" form:"min_amount"`
	MaxAmount    *float64   `json:"max_amount" form:"max_amount"`
	
	// Pagination
	Page      int    `json:"page" form:"page" validate:"min=1"`
	Limit     int    `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy    string `json:"sort_by" form:"sort_by" validate:"oneof=created_at updated_at invoice_date due_date amount"`
	SortOrder string `json:"sort_order" form:"sort_order" validate:"oneof=asc desc"`
}

// OptimizedPaymentQueries provides optimized database queries for payment operations
type OptimizedPaymentQueries struct {
	db *gorm.DB
}

// NewOptimizedPaymentQueries creates a new optimized payment queries service
func NewOptimizedPaymentQueries(db *gorm.DB) *OptimizedPaymentQueries {
	return &OptimizedPaymentQueries{db: db}
}

// GetPaymentsByCompanyIDOptimized gets payments by company with optimized query
func (opq *OptimizedPaymentQueries) GetPaymentsByCompanyIDOptimized(ctx context.Context, companyID string, filters PaymentFilters) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	// Build base query with company filter
	query := opq.db.WithContext(ctx).Model(&models.Payment{}).Where("company_id = ?", companyID)

	// Apply filters with optimized conditions
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.PaymentMethod != nil {
		query = query.Where("payment_method = ?", *filters.PaymentMethod)
	}
	if filters.InvoiceID != nil {
		query = query.Where("invoice_id = ?", *filters.InvoiceID)
	}
	if filters.StartDate != nil {
		query = query.Where("payment_date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("payment_date <= ?", *filters.EndDate)
	}
	if filters.MinAmount != nil {
		query = query.Where("amount >= ?", *filters.MinAmount)
	}
	if filters.MaxAmount != nil {
		query = query.Where("amount <= ?", *filters.MaxAmount)
	}

	// Get total count with optimized query
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Apply sorting with index-friendly ordering
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "payment_date"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query with selective preloading
	if err := query.Preload("Invoice", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, invoice_number, amount, status")
	}).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}

	return payments, total, nil
}

// GetPaymentsByStatusOptimized gets payments by status with optimized query
func (opq *OptimizedPaymentQueries) GetPaymentsByStatusOptimized(ctx context.Context, companyID string, status string) ([]*models.Payment, error) {
	var payments []*models.Payment
	
	// Use composite index on (company_id, status)
	if err := opq.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, status).
		Order("payment_date DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by status: %w", err)
	}
	
	return payments, nil
}

// GetPendingPaymentsOptimized gets pending payments with optimized query
func (opq *OptimizedPaymentQueries) GetPendingPaymentsOptimized(ctx context.Context, companyID string) ([]*models.Payment, error) {
	var payments []*models.Payment
	
	// Use composite index on (company_id, status)
	if err := opq.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, "pending").
		Order("created_at ASC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}
	
	return payments, nil
}

// GetOverduePaymentsOptimized gets overdue payments with optimized query
func (opq *OptimizedPaymentQueries) GetOverduePaymentsOptimized(ctx context.Context, companyID string) ([]*models.Payment, error) {
	var payments []*models.Payment
	now := time.Now()
	
	// Use optimized query for overdue payments
	if err := opq.db.WithContext(ctx).
		Joins("JOIN invoices ON payments.invoice_id = invoices.id").
		Where("payments.company_id = ? AND payments.status = ? AND invoices.due_date < ?", 
			companyID, "pending", now).
		Order("invoices.due_date ASC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue payments: %w", err)
	}
	
	return payments, nil
}

// GetPaymentsByDateRangeOptimized gets payments by date range with optimized query
func (opq *OptimizedPaymentQueries) GetPaymentsByDateRangeOptimized(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.Payment, error) {
	var payments []*models.Payment
	
	// Use index on payment_date with range condition
	if err := opq.db.WithContext(ctx).
		Where("company_id = ? AND payment_date BETWEEN ? AND ?", companyID, startDate, endDate).
		Order("payment_date DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by date range: %w", err)
	}
	
	return payments, nil
}

// GetTotalRevenueOptimized calculates total revenue with optimized query
func (opq *OptimizedPaymentQueries) GetTotalRevenueOptimized(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error) {
	var result struct {
		TotalRevenue float64 `gorm:"column:total_revenue"`
	}
	
	// Use aggregate function for efficient calculation
	if err := opq.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total_revenue").
		Where("company_id = ? AND status = ? AND payment_date BETWEEN ? AND ?", 
			companyID, "completed", startDate, endDate).
		Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}
	
	return result.TotalRevenue, nil
}

// GetPaymentMethodsOptimized gets payment methods breakdown with optimized query
func (opq *OptimizedPaymentQueries) GetPaymentMethodsOptimized(ctx context.Context, companyID string, startDate, endDate time.Time) (map[string]float64, error) {
	var results []struct {
		PaymentMethod string  `gorm:"column:payment_method"`
		TotalAmount   float64 `gorm:"column:total_amount"`
	}
	
	// Use GROUP BY for efficient aggregation
	if err := opq.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("payment_method, COALESCE(SUM(amount), 0) as total_amount").
		Where("company_id = ? AND status = ? AND payment_date BETWEEN ? AND ?", 
			companyID, "completed", startDate, endDate).
		Group("payment_method").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment methods: %w", err)
	}
	
	// Convert to map
	paymentMethods := make(map[string]float64)
	for _, result := range results {
		paymentMethods[result.PaymentMethod] = result.TotalAmount
	}
	
	return paymentMethods, nil
}

// GetInvoicesByCompanyIDOptimized gets invoices by company with optimized query
func (opq *OptimizedPaymentQueries) GetInvoicesByCompanyIDOptimized(ctx context.Context, companyID string, filters InvoiceFilters) ([]*models.Invoice, int64, error) {
	var invoices []*models.Invoice
	var total int64

	// Build base query with company filter
	query := opq.db.WithContext(ctx).Model(&models.Invoice{}).Where("company_id = ?", companyID)

	// Apply filters with optimized conditions
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.StartDate != nil {
		query = query.Where("invoice_date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("invoice_date <= ?", *filters.EndDate)
	}
	if filters.DueDateStart != nil {
		query = query.Where("due_date >= ?", *filters.DueDateStart)
	}
	if filters.DueDateEnd != nil {
		query = query.Where("due_date <= ?", *filters.DueDateEnd)
	}
	if filters.MinAmount != nil {
		query = query.Where("amount >= ?", *filters.MinAmount)
	}
	if filters.MaxAmount != nil {
		query = query.Where("amount <= ?", *filters.MaxAmount)
	}

	// Get total count with optimized query
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Apply sorting with index-friendly ordering
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "invoice_date"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query
	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}

	return invoices, total, nil
}

// GetInvoicesByStatusOptimized gets invoices by status with optimized query
func (opq *OptimizedPaymentQueries) GetInvoicesByStatusOptimized(ctx context.Context, companyID string, status string) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	
	// Use composite index on (company_id, status)
	if err := opq.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, status).
		Order("invoice_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices by status: %w", err)
	}
	
	return invoices, nil
}

// GetOverdueInvoicesOptimized gets overdue invoices with optimized query
func (opq *OptimizedPaymentQueries) GetOverdueInvoicesOptimized(ctx context.Context, companyID string) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	now := time.Now()
	
	// Use index on due_date with optimized condition
	if err := opq.db.WithContext(ctx).
		Where("company_id = ? AND status = ? AND due_date < ?", 
			companyID, "pending", now).
		Order("due_date ASC").
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	
	return invoices, nil
}

// GetInvoiceByNumberOptimized gets invoice by number with optimized query
func (opq *OptimizedPaymentQueries) GetInvoiceByNumberOptimized(ctx context.Context, companyID string, invoiceNumber string) (*models.Invoice, error) {
	var invoice models.Invoice
	
	// Use index on invoice_number
	if err := opq.db.WithContext(ctx).
		Where("company_id = ? AND invoice_number = ?", companyID, invoiceNumber).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice with number %s not found", invoiceNumber)
		}
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}
	
	return &invoice, nil
}

// UpdateInvoiceStatusOptimized updates invoice status with optimized query
func (opq *OptimizedPaymentQueries) UpdateInvoiceStatusOptimized(ctx context.Context, invoiceID string, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	// Use direct update for better performance
	if err := opq.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	
	return nil
}

// GetInvoiceNumberOptimized generates next invoice number with optimized query
func (opq *OptimizedPaymentQueries) GetInvoiceNumberOptimized(ctx context.Context, companyID string, year int) (string, error) {
	var result struct {
		LastNumber int `gorm:"column:last_number"`
	}
	
	// Use optimized query to get the last invoice number
	query := `
		SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM '[0-9]+$') AS INTEGER)), 0) as last_number
		FROM invoices 
		WHERE company_id = ? AND EXTRACT(YEAR FROM invoice_date) = ?
	`
	
	if err := opq.db.WithContext(ctx).Raw(query, companyID, year).Scan(&result).Error; err != nil {
		return "", fmt.Errorf("failed to get invoice number: %w", err)
	}
	
	// Generate next invoice number
	nextNumber := result.LastNumber + 1
	invoiceNumber := fmt.Sprintf("INV-%d-%06d", year, nextNumber)
	
	return invoiceNumber, nil
}

// GetPaymentStatsOptimized gets payment statistics with optimized query
func (opq *OptimizedPaymentQueries) GetPaymentStatsOptimized(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalPayments     int64   `gorm:"column:total_payments"`
		CompletedPayments int64   `gorm:"column:completed_payments"`
		PendingPayments   int64   `gorm:"column:pending_payments"`
		FailedPayments    int64   `gorm:"column:failed_payments"`
		TotalRevenue      float64 `gorm:"column:total_revenue"`
		PendingAmount     float64 `gorm:"column:pending_amount"`
		OverdueAmount     float64 `gorm:"column:overdue_amount"`
	}
	
	// Use single query with conditional aggregation for better performance
	query := `
		SELECT 
			COUNT(*) as total_payments,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_payments,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_payments,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_payments,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN amount END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount END), 0) as pending_amount
		FROM payments 
		WHERE company_id = ?
	`
	
	if err := opq.db.WithContext(ctx).Raw(query, companyID).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}
	
	// Calculate overdue amount separately
	overdueQuery := `
		SELECT COALESCE(SUM(i.amount), 0) as overdue_amount
		FROM invoices i
		WHERE i.company_id = ? AND i.status = 'pending' AND i.due_date < NOW()
	`
	
	var overdueResult struct {
		OverdueAmount float64 `gorm:"column:overdue_amount"`
	}
	
	if err := opq.db.WithContext(ctx).Raw(overdueQuery, companyID).Scan(&overdueResult).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue amount: %w", err)
	}
	
	stats.OverdueAmount = overdueResult.OverdueAmount
	
	return map[string]interface{}{
		"total_payments":     stats.TotalPayments,
		"completed_payments": stats.CompletedPayments,
		"pending_payments":   stats.PendingPayments,
		"failed_payments":    stats.FailedPayments,
		"total_revenue":      stats.TotalRevenue,
		"pending_amount":     stats.PendingAmount,
		"overdue_amount":     stats.OverdueAmount,
	}, nil
}

// GetMonthlyRevenueOptimized gets monthly revenue breakdown with optimized query
func (opq *OptimizedPaymentQueries) GetMonthlyRevenueOptimized(ctx context.Context, companyID string, year int) (map[string]float64, error) {
	var results []struct {
		Month       int     `gorm:"column:month"`
		TotalAmount float64 `gorm:"column:total_amount"`
	}
	
	// Use GROUP BY for efficient aggregation by month
	query := `
		SELECT 
			EXTRACT(MONTH FROM payment_date) as month,
			COALESCE(SUM(amount), 0) as total_amount
		FROM payments 
		WHERE company_id = ? AND status = 'completed' AND EXTRACT(YEAR FROM payment_date) = ?
		GROUP BY EXTRACT(MONTH FROM payment_date)
		ORDER BY month
	`
	
	if err := opq.db.WithContext(ctx).Raw(query, companyID, year).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	
	// Convert to map with month names
	monthlyRevenue := make(map[string]float64)
	monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", 
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	
	for _, result := range results {
		if result.Month >= 1 && result.Month <= 12 {
			monthlyRevenue[monthNames[result.Month]] = result.TotalAmount
		}
	}
	
	return monthlyRevenue, nil
}

// BatchUpdatePaymentStatusOptimized updates multiple payment statuses in a single transaction
func (opq *OptimizedPaymentQueries) BatchUpdatePaymentStatusOptimized(ctx context.Context, paymentIDs []string, status string, reason string) error {
	if len(paymentIDs) == 0 {
		return nil
	}
	
	// Use batch update for better performance
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	if reason != "" {
		updates["notes"] = reason
	}
	
	if err := opq.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id IN ?", paymentIDs).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to batch update payment status: %w", err)
	}
	
	return nil
}
