package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	redis       *redis.Client
	cfg         *config.Config
	repoManager *repository.RepositoryManager
	cache       *CacheService
}

// CacheService provides caching functionality for payment operations
type CacheService struct {
	redis *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{redis: redis}
}

// GetInvoiceFromCache retrieves an invoice from cache
func (cs *CacheService) GetInvoiceFromCache(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	key := fmt.Sprintf("invoice:%s", invoiceID)
	
	var invoice models.Invoice
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get invoice from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice from cache: %w", err)
	}
	
	return &invoice, nil
}

// SetInvoiceInCache stores an invoice in cache
func (cs *CacheService) SetInvoiceInCache(ctx context.Context, invoice *models.Invoice, expiration time.Duration) error {
	key := fmt.Sprintf("invoice:%s", invoice.ID)
	
	data, err := json.Marshal(invoice)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set invoice in cache: %w", err)
	}
	
	return nil
}

// InvalidateInvoiceCache removes an invoice from cache
func (cs *CacheService) InvalidateInvoiceCache(ctx context.Context, invoiceID string) error {
	key := fmt.Sprintf("invoice:%s", invoiceID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate invoice cache: %w", err)
	}
	
	return nil
}

// GetPaymentFromCache retrieves a payment from cache
func (cs *CacheService) GetPaymentFromCache(ctx context.Context, paymentID string) (*models.Payment, error) {
	key := fmt.Sprintf("payment:%s", paymentID)
	
	var payment models.Payment
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get payment from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &payment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment from cache: %w", err)
	}
	
	return &payment, nil
}

// SetPaymentInCache stores a payment in cache
func (cs *CacheService) SetPaymentInCache(ctx context.Context, payment *models.Payment, expiration time.Duration) error {
	key := fmt.Sprintf("payment:%s", payment.ID)
	
	data, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set payment in cache: %w", err)
	}
	
	return nil
}

// InvalidatePaymentCache removes a payment from cache
func (cs *CacheService) InvalidatePaymentCache(ctx context.Context, paymentID string) error {
	key := fmt.Sprintf("payment:%s", paymentID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate payment cache: %w", err)
	}
	
	return nil
}

// GetInvoiceListFromCache retrieves an invoice list from cache
func (cs *CacheService) GetInvoiceListFromCache(ctx context.Context, companyID string, status string, limit, offset int) ([]*models.Invoice, error) {
	cacheKey := cs.generateInvoiceListCacheKey(companyID, status, limit, offset)
	
	var invoices []*models.Invoice
	data, err := cs.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get invoice list from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &invoices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice list from cache: %w", err)
	}
	
	return invoices, nil
}

// SetInvoiceListInCache stores an invoice list in cache
func (cs *CacheService) SetInvoiceListInCache(ctx context.Context, companyID string, status string, limit, offset int, invoices []*models.Invoice, expiration time.Duration) error {
	cacheKey := cs.generateInvoiceListCacheKey(companyID, status, limit, offset)
	
	data, err := json.Marshal(invoices)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice list for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, cacheKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set invoice list in cache: %w", err)
	}
	
	return nil
}

// InvalidateInvoiceListCache removes invoice list cache for a company
func (cs *CacheService) InvalidateInvoiceListCache(ctx context.Context, companyID string) error {
	pattern := fmt.Sprintf("invoice:list:%s:*", companyID)
	
	keys, err := cs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get invoice list cache keys: %w", err)
	}
	
	if len(keys) > 0 {
		if err := cs.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate invoice list cache: %w", err)
		}
	}
	
	return nil
}

// GetPaymentInstructionsFromCache retrieves payment instructions from cache
func (cs *CacheService) GetPaymentInstructionsFromCache(ctx context.Context, invoiceID string) (*PaymentInstructions, error) {
	key := fmt.Sprintf("payment_instructions:%s", invoiceID)
	
	var instructions PaymentInstructions
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get payment instructions from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &instructions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment instructions from cache: %w", err)
	}
	
	return &instructions, nil
}

// SetPaymentInstructionsInCache stores payment instructions in cache
func (cs *CacheService) SetPaymentInstructionsInCache(ctx context.Context, invoiceID string, instructions *PaymentInstructions, expiration time.Duration) error {
	key := fmt.Sprintf("payment_instructions:%s", invoiceID)
	
	data, err := json.Marshal(instructions)
	if err != nil {
		return fmt.Errorf("failed to marshal payment instructions for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set payment instructions in cache: %w", err)
	}
	
	return nil
}

// InvalidatePaymentInstructionsCache removes payment instructions from cache
func (cs *CacheService) InvalidatePaymentInstructionsCache(ctx context.Context, invoiceID string) error {
	key := fmt.Sprintf("payment_instructions:%s", invoiceID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate payment instructions cache: %w", err)
	}
	
	return nil
}

// generateInvoiceListCacheKey creates a cache key for invoice list queries
func (cs *CacheService) generateInvoiceListCacheKey(companyID string, status string, limit, offset int) string {
	return fmt.Sprintf("invoice:list:%s:%s:%d:%d", companyID, status, limit, offset)
}

// GetInvoiceByNumberFromCache retrieves an invoice by invoice number from cache
func (cs *CacheService) GetInvoiceByNumberFromCache(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	key := fmt.Sprintf("invoice:number:%s", invoiceNumber)
	
	var invoice models.Invoice
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get invoice by number from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice by number from cache: %w", err)
	}
	
	return &invoice, nil
}

// SetInvoiceByNumberInCache stores an invoice by invoice number in cache
func (cs *CacheService) SetInvoiceByNumberInCache(ctx context.Context, invoice *models.Invoice, expiration time.Duration) error {
	key := fmt.Sprintf("invoice:number:%s", invoice.InvoiceNumber)
	
	data, err := json.Marshal(invoice)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice by number for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set invoice by number in cache: %w", err)
	}
	
	return nil
}

// InvalidateInvoiceByNumberCache removes an invoice by number from cache
func (cs *CacheService) InvalidateInvoiceByNumberCache(ctx context.Context, invoiceNumber string) error {
	key := fmt.Sprintf("invoice:number:%s", invoiceNumber)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate invoice by number cache: %w", err)
	}
	
	return nil
}

// GetPaymentByReferenceFromCache retrieves a payment by reference number from cache
func (cs *CacheService) GetPaymentByReferenceFromCache(ctx context.Context, referenceNumber string) (*models.Payment, error) {
	key := fmt.Sprintf("payment:reference:%s", referenceNumber)
	
	var payment models.Payment
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get payment by reference from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &payment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment by reference from cache: %w", err)
	}
	
	return &payment, nil
}

// SetPaymentByReferenceInCache stores a payment by reference number in cache
func (cs *CacheService) SetPaymentByReferenceInCache(ctx context.Context, payment *models.Payment, expiration time.Duration) error {
	key := fmt.Sprintf("payment:reference:%s", payment.ReferenceNumber)
	
	data, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment by reference for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set payment by reference in cache: %w", err)
	}
	
	return nil
}

// InvalidatePaymentByReferenceCache removes a payment by reference from cache
func (cs *CacheService) InvalidatePaymentByReferenceCache(ctx context.Context, referenceNumber string) error {
	key := fmt.Sprintf("payment:reference:%s", referenceNumber)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate payment by reference cache: %w", err)
	}
	
	return nil
}

// BulkSetInvoicesInCache stores multiple invoices in cache
func (cs *CacheService) BulkSetInvoicesInCache(ctx context.Context, invoices []*models.Invoice, expiration time.Duration) error {
	if len(invoices) == 0 {
		return nil
	}
	
	pipe := cs.redis.Pipeline()
	
	for _, invoice := range invoices {
		// Cache by ID
		idKey := fmt.Sprintf("invoice:%s", invoice.ID)
		idData, err := json.Marshal(invoice)
		if err != nil {
			return fmt.Errorf("failed to marshal invoice %s for cache: %w", invoice.ID, err)
		}
		pipe.Set(ctx, idKey, idData, expiration)
		
		// Cache by invoice number
		numberKey := fmt.Sprintf("invoice:number:%s", invoice.InvoiceNumber)
		pipe.Set(ctx, numberKey, idData, expiration)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to bulk set invoices in cache: %w", err)
	}
	
	return nil
}

// BulkSetPaymentsInCache stores multiple payments in cache
func (cs *CacheService) BulkSetPaymentsInCache(ctx context.Context, payments []*models.Payment, expiration time.Duration) error {
	if len(payments) == 0 {
		return nil
	}
	
	pipe := cs.redis.Pipeline()
	
	for _, payment := range payments {
		// Cache by ID
		idKey := fmt.Sprintf("payment:%s", payment.ID)
		idData, err := json.Marshal(payment)
		if err != nil {
			return fmt.Errorf("failed to marshal payment %s for cache: %w", payment.ID, err)
		}
		pipe.Set(ctx, idKey, idData, expiration)
		
		// Cache by reference number if available
		if payment.ReferenceNumber != "" {
			refKey := fmt.Sprintf("payment:reference:%s", payment.ReferenceNumber)
			pipe.Set(ctx, refKey, idData, expiration)
		}
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to bulk set payments in cache: %w", err)
	}
	
	return nil
}

func NewService(db *gorm.DB, redis *redis.Client, cfg *config.Config, repoManager *repository.RepositoryManager) *Service {
	return &Service{
		db:          db,
		redis:       redis,
		cfg:         cfg,
		repoManager: repoManager,
		cache:       NewCacheService(redis),
	}
}

// InvoiceRequest represents a request to create an invoice
type InvoiceRequest struct {
	CompanyID      string    `json:"company_id" binding:"required"`
	SubscriptionID string    `json:"subscription_id"`
	BillingPeriod  string    `json:"billing_period" binding:"required"`
	DueDate        time.Time `json:"due_date"`
	Notes          string    `json:"notes"`
}

// InvoiceResponse represents the response for invoice creation
type InvoiceResponse struct {
	InvoiceID             string                 `json:"invoice_id"`
	InvoiceNumber         string                 `json:"invoice_number"`
	Subtotal              float64                `json:"subtotal"`
	TaxAmount             float64                `json:"tax_amount"`
	TotalAmount           float64                `json:"total_amount"`
	DueDate               string                 `json:"due_date"`
	PaymentInstructions   PaymentInstructions    `json:"payment_instructions"`
	InvoicePDF            string                 `json:"invoice_pdf,omitempty"` // Base64 encoded PDF
}

// PaymentInstructions contains bank transfer instructions
type PaymentInstructions struct {
	BankName        string `json:"bank_name"`
	AccountNumber   string `json:"account_number"`
	AccountHolder   string `json:"account_holder"`
	ReferenceCode   string `json:"reference_code"`
	Amount          string `json:"amount"`
	TransferNote    string `json:"transfer_note"`
}

// PaymentConfirmationRequest represents a payment confirmation
type PaymentConfirmationRequest struct {
	InvoiceID        string    `json:"invoice_id" binding:"required"`
	BankAccount      string    `json:"bank_account" binding:"required"`
	TransferAmount   float64   `json:"transfer_amount" binding:"required"`
	TransferDate     time.Time `json:"transfer_date" binding:"required"`
	ReferenceNumber  string    `json:"reference_number"`
	Notes            string    `json:"notes"`
}

// SubscriptionBillingRequest represents subscription billing request
type SubscriptionBillingRequest struct {
	CompanyID        string `json:"company_id" binding:"required"`
	SubscriptionID   string `json:"subscription_id" binding:"required"`
	BillingCycle     string `json:"billing_cycle" binding:"required"` // monthly, yearly
	StartDate        string `json:"start_date" binding:"required"`
	EndDate          string `json:"end_date" binding:"required"`
}

// GenerateInvoice creates a new invoice with Indonesian compliance
func (s *Service) GenerateInvoice(ctx context.Context, req *InvoiceRequest) (*InvoiceResponse, error) {
	// Get company details for NPWP and tax information
	company, err := s.repoManager.GetCompanies().GetByID(ctx, req.CompanyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("company")
		}
		return nil, apperrors.Wrap(err, "failed to get company")
	}

	// Generate invoice number (Indonesian format: INV/YYYY/MM/XXXX)
	invoiceNumber := s.generateInvoiceNumber(req.CompanyID)

	// Calculate billing period
	startDate, endDate, err := s.parseBillingPeriod(req.BillingPeriod)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to parse billing period")
	}

	// Calculate amounts (assuming subscription-based billing)
	subtotal := 250000.0 // Base subscription amount in IDR
	taxRate := 11.0      // Indonesian PPN rate
	taxAmount := subtotal * (taxRate / 100)
	totalAmount := subtotal + taxAmount

	// Set due date (default to 14 days from invoice date)
	dueDate := req.DueDate
	if dueDate.IsZero() {
		dueDate = time.Now().AddDate(0, 0, 14)
	}

	// Create invoice
	invoice := &models.Invoice{
		CompanyID:          req.CompanyID,
		SubscriptionID:     &req.SubscriptionID,
		InvoiceNumber:      invoiceNumber,
		InvoiceDate:        time.Now(),
		DueDate:            dueDate,
		BillingPeriodStart: startDate,
		BillingPeriodEnd:   endDate,
		Subtotal:           subtotal,
		TaxAmount:          taxAmount,
		TotalAmount:        totalAmount,
		BalanceAmount:      totalAmount,
		Status:             "draft",
		TaxNumber:          company.NPWP,
		TaxRate:            taxRate,
		Currency:           "IDR",
		Notes:              req.Notes,
		Terms:              "Payment due within 14 days of invoice date. Late payments may incur additional charges.",
	}

	// Create invoice items
	invoice.Items = models.JSON{
		"items": []map[string]interface{}{
			{
				"description": "FleetTracker Pro Subscription",
				"quantity":    1,
				"unit_price":  subtotal,
				"total":       subtotal,
			},
		},
	}

	// Save invoice to database
	if err := s.repoManager.InvoiceRepository().Create(ctx, invoice); err != nil {
		return nil, apperrors.Wrap(err, "failed to create invoice")
	}

	// Invalidate invoice list cache after creating new invoice
	if err := s.cache.InvalidateInvoiceListCache(ctx, req.CompanyID); err != nil {
		// Log cache invalidation error but don't fail the request
		fmt.Printf("Failed to invalidate invoice list cache %s: %v\n", req.CompanyID, err)
	}

	// Generate payment instructions
	paymentInstructions := s.generatePaymentInstructions(invoice, company)

	// Generate PDF (placeholder - would use actual PDF generation library)
	pdfContent := s.generateInvoicePDF(invoice, company)

	response := &InvoiceResponse{
		InvoiceID:           invoice.ID,
		InvoiceNumber:       invoice.InvoiceNumber,
		Subtotal:            invoice.Subtotal,
		TaxAmount:           invoice.TaxAmount,
		TotalAmount:         invoice.TotalAmount,
		DueDate:             invoice.DueDate.Format("2006-01-02"),
		PaymentInstructions: paymentInstructions,
		InvoicePDF:          pdfContent,
	}

	return response, nil
}

// ConfirmPayment confirms a manual bank transfer payment
func (s *Service) ConfirmPayment(ctx context.Context, req *PaymentConfirmationRequest, confirmedBy string) error {
	// Get invoice using cached method
	invoice, err := s.GetInvoice(ctx, req.InvoiceID)
	if err != nil {
		return err
	}

	// Validate payment amount
	if req.TransferAmount != invoice.TotalAmount {
		return apperrors.NewBadRequestError(fmt.Sprintf("payment amount mismatch: expected %.2f, got %.2f", invoice.TotalAmount, req.TransferAmount))
	}

	// Create payment record
	payment := &models.Payment{
		CompanyID:       invoice.CompanyID,
		SubscriptionID:  invoice.SubscriptionID,
		Amount:          invoice.Subtotal,
		TaxAmount:       invoice.TaxAmount,
		TotalAmount:     invoice.TotalAmount,
		PaymentMethod:   "bank_transfer",
		PaymentType:     "subscription",
		Status:          "completed",
		ReferenceNumber: req.ReferenceNumber,
		CompletedAt:     &req.TransferDate,
	}

	// Set bank transfer data
	payment.SetBankTransferData(map[string]interface{}{
		"bank_account":      req.BankAccount,
		"transfer_amount":   req.TransferAmount,
		"transfer_date":     req.TransferDate,
		"reference_number":  req.ReferenceNumber,
		"confirmed_by":      confirmedBy,
		"notes":            req.Notes,
	})

	// Save payment
	if err := s.repoManager.PaymentRepository().Create(ctx, payment); err != nil {
		return apperrors.Wrap(err, "failed to create payment")
	}

	// Update invoice
	invoice.MarkAsPaid(req.TransferAmount, "bank_transfer")
	invoice.PaymentReference = req.ReferenceNumber
	invoice.PaymentID = &payment.ID

	if err := s.repoManager.InvoiceRepository().Update(ctx, invoice); err != nil {
		return apperrors.Wrap(err, "failed to update invoice")
	}

	// Invalidate caches after payment confirmation
	if err := s.cache.InvalidateInvoiceCache(ctx, invoice.ID); err != nil {
		fmt.Printf("Failed to invalidate invoice cache %s: %v\n", invoice.ID, err)
	}
	
	if err := s.cache.InvalidateInvoiceByNumberCache(ctx, invoice.InvoiceNumber); err != nil {
		fmt.Printf("Failed to invalidate invoice by number cache %s: %v\n", invoice.InvoiceNumber, err)
	}
	
	if err := s.cache.InvalidateInvoiceListCache(ctx, invoice.CompanyID); err != nil {
		fmt.Printf("Failed to invalidate invoice list cache %s: %v\n", invoice.CompanyID, err)
	}
	
	if err := s.cache.InvalidatePaymentInstructionsCache(ctx, invoice.ID); err != nil {
		fmt.Printf("Failed to invalidate payment instructions cache %s: %v\n", invoice.ID, err)
	}
	
	// Invalidate payment caches if payment was created
	if payment.ReferenceNumber != "" {
		if err := s.cache.InvalidatePaymentByReferenceCache(ctx, payment.ReferenceNumber); err != nil {
			fmt.Printf("Failed to invalidate payment by reference cache %s: %v\n", payment.ReferenceNumber, err)
		}
	}

	return nil
}

// GetInvoice retrieves an individual invoice by ID with caching
func (s *Service) GetInvoice(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	// Try to get from cache first
	cachedInvoice, err := s.cache.GetInvoiceFromCache(ctx, invoiceID)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for invoice %s: %v\n", invoiceID, err)
	}
	
	if cachedInvoice != nil {
		return cachedInvoice, nil
	}
	
	// Get from database
	invoice, err := s.repoManager.GetInvoices().GetByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("invoice")
		}
		return nil, apperrors.Wrap(err, "failed to get invoice")
	}

	// Cache the result for 30 minutes
	if err := s.cache.SetInvoiceInCache(ctx, invoice, 30*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache invoice %s: %v\n", invoiceID, err)
	}
	
	// Also cache by invoice number
	if err := s.cache.SetInvoiceByNumberInCache(ctx, invoice, 30*time.Minute); err != nil {
		fmt.Printf("Failed to cache invoice by number %s: %v\n", invoice.InvoiceNumber, err)
	}

	return invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number with caching
func (s *Service) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	// Try to get from cache first
	cachedInvoice, err := s.cache.GetInvoiceByNumberFromCache(ctx, invoiceNumber)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for invoice number %s: %v\n", invoiceNumber, err)
	}
	
	if cachedInvoice != nil {
		return cachedInvoice, nil
	}
	
	// Get from database (assuming there's a method to get by invoice number)
	// For now, we'll use a simple query - in production, you'd have a proper repository method
	var invoice models.Invoice
	if err := s.db.Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("invoice")
		}
		return nil, apperrors.Wrap(err, "failed to get invoice by number")
	}

	// Cache the result for 30 minutes
	if err := s.cache.SetInvoiceInCache(ctx, &invoice, 30*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache invoice %s: %v\n", invoice.ID, err)
	}
	
	// Also cache by invoice number
	if err := s.cache.SetInvoiceByNumberInCache(ctx, &invoice, 30*time.Minute); err != nil {
		fmt.Printf("Failed to cache invoice by number %s: %v\n", invoice.InvoiceNumber, err)
	}

	return &invoice, nil
}

// GetPayment retrieves an individual payment by ID with caching
func (s *Service) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	// Try to get from cache first
	cachedPayment, err := s.cache.GetPaymentFromCache(ctx, paymentID)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for payment %s: %v\n", paymentID, err)
	}
	
	if cachedPayment != nil {
		return cachedPayment, nil
	}
	
	// Get from database
	payment, err := s.repoManager.PaymentRepository().GetByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("payment")
		}
		return nil, apperrors.Wrap(err, "failed to get payment")
	}

	// Cache the result for 30 minutes
	if err := s.cache.SetPaymentInCache(ctx, payment, 30*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache payment %s: %v\n", paymentID, err)
	}
	
	// Also cache by reference number if available
	if payment.ReferenceNumber != "" {
		if err := s.cache.SetPaymentByReferenceInCache(ctx, payment, 30*time.Minute); err != nil {
			fmt.Printf("Failed to cache payment by reference %s: %v\n", payment.ReferenceNumber, err)
		}
	}

	return payment, nil
}

// GetPaymentByReference retrieves a payment by reference number with caching
func (s *Service) GetPaymentByReference(ctx context.Context, referenceNumber string) (*models.Payment, error) {
	// Try to get from cache first
	cachedPayment, err := s.cache.GetPaymentByReferenceFromCache(ctx, referenceNumber)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for payment reference %s: %v\n", referenceNumber, err)
	}
	
	if cachedPayment != nil {
		return cachedPayment, nil
	}
	
	// Get from database
	var payment models.Payment
	if err := s.db.Where("reference_number = ?", referenceNumber).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("payment")
		}
		return nil, apperrors.Wrap(err, "failed to get payment by reference")
	}

	// Cache the result for 30 minutes
	if err := s.cache.SetPaymentInCache(ctx, &payment, 30*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache payment %s: %v\n", payment.ID, err)
	}
	
	// Also cache by reference number
	if err := s.cache.SetPaymentByReferenceInCache(ctx, &payment, 30*time.Minute); err != nil {
		fmt.Printf("Failed to cache payment by reference %s: %v\n", payment.ReferenceNumber, err)
	}

	return &payment, nil
}

// GetInvoices retrieves invoices for a company
func (s *Service) GetInvoices(ctx context.Context, companyID string, status string, limit, offset int) ([]*models.Invoice, error) {
	// Try to get from cache first
	cachedInvoices, err := s.cache.GetInvoiceListFromCache(ctx, companyID, status, limit, offset)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for invoice list %s: %v\n", companyID, err)
	}
	
	if cachedInvoices != nil {
		return cachedInvoices, nil
	}
	
	filters := map[string]interface{}{
		"company_id": companyID,
	}
	if status != "" {
		filters["status"] = status
	}

	// Create pagination and filter options
	pagination := repository.Pagination{
		Page:     1,
		PageSize: limit,
		Offset:   offset,
		Limit:    limit,
	}
	
	filterOptions := repository.FilterOptions{
		CompanyID: companyID,
		Conditions: []repository.Condition{
			{Field: "company_id", Operator: "=", Value: companyID},
		},
	}
	
	invoices, err := s.repoManager.GetInvoices().List(ctx, filterOptions, pagination)
	if err != nil {
		return nil, err
	}

	// Cache the result for 10 minutes (shorter TTL for payment data)
	if err := s.cache.SetInvoiceListInCache(ctx, companyID, status, limit, offset, invoices, 10*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache invoice list %s: %v\n", companyID, err)
	}

	// Also bulk cache individual invoices for faster individual access
	if err := s.cache.BulkSetInvoicesInCache(ctx, invoices, 30*time.Minute); err != nil {
		fmt.Printf("Failed to bulk cache invoices for %s: %v\n", companyID, err)
	}

	return invoices, nil
}

// GetPaymentInstructions generates payment instructions for an invoice
func (s *Service) GetPaymentInstructions(ctx context.Context, invoiceID string) (*PaymentInstructions, error) {
	// Try to get from cache first
	cachedInstructions, err := s.cache.GetPaymentInstructionsFromCache(ctx, invoiceID)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for payment instructions %s: %v\n", invoiceID, err)
	}
	
	if cachedInstructions != nil {
		return cachedInstructions, nil
	}
	
	invoice, err := s.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	company, err := s.repoManager.GetCompanies().GetByID(ctx, invoice.CompanyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("company")
		}
		return nil, apperrors.Wrap(err, "failed to get company")
	}

	instructions := s.generatePaymentInstructions(invoice, company)
	
	// Cache the result for 1 hour (payment instructions don't change often)
	if err := s.cache.SetPaymentInstructionsInCache(ctx, invoiceID, &instructions, 1*time.Hour); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache payment instructions %s: %v\n", invoiceID, err)
	}
	
	return &instructions, nil
}

// GenerateSubscriptionBilling creates automatic billing for subscriptions
func (s *Service) GenerateSubscriptionBilling(ctx context.Context, req *SubscriptionBillingRequest) error {
	// Get subscription
	_, err := s.repoManager.GetSubscriptions().GetByID(ctx, req.SubscriptionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("subscription")
		}
		return apperrors.Wrap(err, "failed to get subscription")
	}

	// Parse dates
	_, err = time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return apperrors.NewValidationError("invalid start date format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return apperrors.NewValidationError("invalid end date format")
	}

	// Create invoice request
	invoiceReq := &InvoiceRequest{
		CompanyID:      req.CompanyID,
		SubscriptionID: req.SubscriptionID,
		BillingPeriod:  fmt.Sprintf("%s to %s", req.StartDate, req.EndDate),
		DueDate:        endDate.AddDate(0, 0, 14), // 14 days after billing period
		Notes:          fmt.Sprintf("Automatic billing for %s subscription", req.BillingCycle),
	}

	// Generate invoice
	_, err = s.GenerateInvoice(ctx, invoiceReq)
	if err != nil {
		return apperrors.Wrap(err, "failed to generate subscription billing")
	}

	return nil
}

// Helper methods

// generateInvoiceNumber creates Indonesian-style invoice number
func (s *Service) generateInvoiceNumber(companyID string) string {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	
	// Get next sequence number for this month
	var count int64
	s.db.Model(&models.Invoice{}).
		Where("company_id = ? AND EXTRACT(year FROM created_at) = ? AND EXTRACT(month FROM created_at) = ?", 
			companyID, year, month).
		Count(&count)
	
	sequence := count + 1
	return fmt.Sprintf("INV/%d/%02d/%04d", year, month, sequence)
}

// parseBillingPeriod parses billing period string
func (s *Service) parseBillingPeriod(_ string) (time.Time, time.Time, error) {
	// Simple parsing - can be enhanced
	now := time.Now()
	startDate := now.AddDate(0, -1, 0) // Last month
	endDate := now
	return startDate, endDate, nil
}

// generatePaymentInstructions creates bank transfer instructions
func (s *Service) generatePaymentInstructions(invoice *models.Invoice, company *models.Company) PaymentInstructions {
	// Use company's primary bank account or default
	bankAccount := "1234567890" // Default bank account
	bankName := "Bank Central Asia (BCA)"
	accountHolder := "PT FleetTracker Indonesia"
	
	// Generate reference code
	referenceCode := fmt.Sprintf("INV%s", invoice.InvoiceNumber)

	return PaymentInstructions{
		BankName:      bankName,
		AccountNumber: bankAccount,
		AccountHolder: accountHolder,
		ReferenceCode: referenceCode,
		Amount:        fmt.Sprintf("Rp %s", formatNumber(invoice.TotalAmount)),
		TransferNote:  fmt.Sprintf("Payment for invoice %s - %s", invoice.InvoiceNumber, company.Name),
	}
}

// generateInvoicePDF creates PDF content (placeholder)
func (s *Service) generateInvoicePDF(_ *models.Invoice, _ *models.Company) string {
	// This would use a PDF generation library like gofpdf
	// For now, return a placeholder
	return "PDF_CONTENT_PLACEHOLDER"
}

// formatNumber formats number with Indonesian formatting
func formatNumber(num float64) string {
	return fmt.Sprintf("%.2f", num)
}
