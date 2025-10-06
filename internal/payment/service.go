package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	cfg         *config.Config
	repoManager *repository.RepositoryManager
}

func NewService(db *gorm.DB, cfg *config.Config, repoManager *repository.RepositoryManager) *Service {
	return &Service{
		db:          db,
		cfg:         cfg,
		repoManager: repoManager,
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
	// Get invoice
	invoice, err := s.repoManager.GetInvoices().GetByID(ctx, req.InvoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("invoice")
		}
		return apperrors.Wrap(err, "failed to get invoice")
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

	return nil
}

// GetInvoices retrieves invoices for a company
func (s *Service) GetInvoices(ctx context.Context, companyID string, status string, limit, offset int) ([]*models.Invoice, error) {
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
	
	return s.repoManager.GetInvoices().List(ctx, filterOptions, pagination)
}

// GetPaymentInstructions generates payment instructions for an invoice
func (s *Service) GetPaymentInstructions(ctx context.Context, invoiceID string) (*PaymentInstructions, error) {
	invoice, err := s.repoManager.GetInvoices().GetByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("invoice")
		}
		return nil, apperrors.Wrap(err, "failed to get invoice")
	}

	company, err := s.repoManager.GetCompanies().GetByID(ctx, invoice.CompanyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("company")
		}
		return nil, apperrors.Wrap(err, "failed to get company")
	}

	instructions := s.generatePaymentInstructions(invoice, company)
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
