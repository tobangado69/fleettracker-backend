package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Subscription represents a company's subscription plan
type Subscription struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Subscription Plan
	PlanName        string    `json:"plan_name" gorm:"type:varchar(100);not null"` // basic, premium, enterprise
	PlanType        string    `json:"plan_type" gorm:"type:varchar(50);not null"` // monthly, yearly
	MaxVehicles     int       `json:"max_vehicles" gorm:"not null"`
	MaxDrivers      int       `json:"max_drivers" gorm:"not null"`
	MaxUsers        int       `json:"max_users" gorm:"not null"`
	
	// Pricing (Indonesian Rupiah)
	Price           float64   `json:"price" gorm:"type:decimal(12,2);not null"` // in IDR
	Currency        string    `json:"currency" gorm:"type:varchar(3);default:'IDR'"`
	TaxRate         float64   `json:"tax_rate" gorm:"type:decimal(5,2);default:11.00"` // Indonesian VAT rate
	
	// Subscription Status
	Status          string    `json:"status" gorm:"type:varchar(20);default:'active'"` // active, suspended, cancelled, expired
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	
	// Billing Period
	StartDate       time.Time `json:"start_date" gorm:"not null"`
	EndDate         time.Time `json:"end_date" gorm:"not null"`
	NextBillingDate time.Time `json:"next_billing_date"`
	
	// Auto-renewal
	AutoRenew       bool      `json:"auto_renew" gorm:"default:true"`
	RenewalPeriod   int       `json:"renewal_period" gorm:"default:1"` // months
	
	// Features
	Features        JSON      `json:"features" gorm:"type:jsonb"` // Plan features and limits
	
	// Timestamps
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Payments []Payment `json:"payments,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// Payment represents a payment transaction
type Payment struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID      string    `json:"company_id" gorm:"type:uuid;not null;index"`
	SubscriptionID *string   `json:"subscription_id" gorm:"type:uuid;index"`
	
	// Payment Information
	Amount         float64   `json:"amount" gorm:"type:decimal(12,2);not null"` // in IDR
	Currency       string    `json:"currency" gorm:"type:varchar(3);default:'IDR'"`
	TaxAmount      float64   `json:"tax_amount" gorm:"type:decimal(12,2);default:0"` // in IDR
	TotalAmount    float64   `json:"total_amount" gorm:"type:decimal(12,2);not null"` // amount + tax
	
	// Payment Method
	PaymentMethod  string    `json:"payment_method" gorm:"type:varchar(50);not null"` // qris, bank_transfer, e_wallet, credit_card
	PaymentType    string    `json:"payment_type" gorm:"type:varchar(50);not null"` // subscription, invoice, top_up
	
	// Payment Status
	Status         string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, processing, completed, failed, cancelled, refunded
	
	// Indonesian Payment Details
	QRISData       JSON      `json:"qris_data" gorm:"type:jsonb"` // QRIS payment details
	BankTransfer   JSON      `json:"bank_transfer" gorm:"type:jsonb"` // Bank transfer details
	EWalletData    JSON      `json:"e_wallet_data" gorm:"type:jsonb"` // E-wallet payment details
	
	// Transaction References
	ExternalID     string    `json:"external_id" gorm:"type:varchar(255)"` // External payment gateway ID
	ReferenceNumber string   `json:"reference_number" gorm:"type:varchar(100)"` // Internal reference
	TransactionID  string    `json:"transaction_id" gorm:"type:varchar(255)"` // Payment gateway transaction ID
	
	// Payment Gateway Response
	GatewayResponse JSON     `json:"gateway_response" gorm:"type:jsonb"` // Full gateway response
	CallbackData    JSON     `json:"callback_data" gorm:"type:jsonb"` // Payment callback data
	
	// Timestamps
	InitiatedAt     time.Time `json:"initiated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	Company      Company       `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
	Invoice      *Invoice      `json:"invoice,omitempty" gorm:"foreignKey:PaymentID"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID      string    `json:"company_id" gorm:"type:uuid;not null;index"`
	SubscriptionID *string   `json:"subscription_id" gorm:"type:uuid;index"`
	PaymentID      *string   `json:"payment_id" gorm:"type:uuid;index"`
	
	// Invoice Information
	InvoiceNumber  string    `json:"invoice_number" gorm:"type:varchar(50);unique;not null"`
	InvoiceDate    time.Time `json:"invoice_date" gorm:"not null"`
	DueDate        time.Time `json:"due_date" gorm:"not null"`
	
	// Billing Period
	BillingPeriodStart time.Time `json:"billing_period_start" gorm:"not null"`
	BillingPeriodEnd   time.Time `json:"billing_period_end" gorm:"not null"`
	
	// Amount Details
	Subtotal       float64   `json:"subtotal" gorm:"type:decimal(12,2);not null"` // in IDR
	TaxAmount      float64   `json:"tax_amount" gorm:"type:decimal(12,2);default:0"` // in IDR
	TotalAmount    float64   `json:"total_amount" gorm:"type:decimal(12,2);not null"` // in IDR
	PaidAmount     float64   `json:"paid_amount" gorm:"type:decimal(12,2);default:0"` // in IDR
	BalanceAmount  float64   `json:"balance_amount" gorm:"type:decimal(12,2)"` // in IDR
	
	// Invoice Status
	Status         string    `json:"status" gorm:"type:varchar(20);default:'draft'"` // draft, sent, paid, overdue, cancelled
	
	// Indonesian Tax Information
	TaxNumber      string    `json:"tax_number" gorm:"type:varchar(20)"` // Company tax number
	TaxRate        float64   `json:"tax_rate" gorm:"type:decimal(5,2);default:11.00"` // Indonesian VAT rate
	Currency       string    `json:"currency" gorm:"type:varchar(3);default:'IDR'"`
	
	// Invoice Items
	Items          JSON      `json:"items" gorm:"type:jsonb"` // Invoice line items
	
	// Payment Information
	PaymentMethod  string    `json:"payment_method" gorm:"type:varchar(50)"` // How it was paid
	PaymentReference string  `json:"payment_reference" gorm:"type:varchar(255)"` // Payment reference
	
	// Notes and Terms
	Notes          string    `json:"notes" gorm:"type:text"`
	Terms          string    `json:"terms" gorm:"type:text"`
	
	// Timestamps
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Company      Company       `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
	Payment      *Payment      `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
}

// TableName specifies the table name for the Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}

// TableName specifies the table name for the Payment model
func (Payment) TableName() string {
	return "payments"
}

// TableName specifies the table name for the Invoice model
func (Invoice) TableName() string {
	return "invoices"
}

// BeforeCreate hook for Subscription
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.Currency == "" {
		s.Currency = "IDR"
	}
	if s.Status == "" {
		s.Status = "active"
	}
	if s.TaxRate == 0 {
		s.TaxRate = 11.00 // Indonesian VAT rate
	}
	return nil
}

// BeforeCreate hook for Payment
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.Currency == "" {
		p.Currency = "IDR"
	}
	if p.Status == "" {
		p.Status = "pending"
	}
	p.TotalAmount = p.Amount + p.TaxAmount
	p.InitiatedAt = time.Now()
	return nil
}

// BeforeCreate hook for Invoice
func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.Status == "" {
		i.Status = "draft"
	}
	if i.TaxRate == 0 {
		i.TaxRate = 11.00 // Indonesian VAT rate
	}
	if i.Currency == "" {
		i.Currency = "IDR"
	}
	i.TotalAmount = i.Subtotal + i.TaxAmount
	i.BalanceAmount = i.TotalAmount - i.PaidAmount
	return nil
}

// Subscription methods
// IsSubscriptionActive checks if subscription is currently active
func (s *Subscription) IsSubscriptionActive() bool {
	return s.IsActive && s.Status == "active" && time.Now().Before(s.EndDate)
}

// IsExpired checks if subscription is expired
func (s *Subscription) IsExpired() bool {
	return time.Now().After(s.EndDate)
}

// DaysUntilExpiry returns days until subscription expires
func (s *Subscription) DaysUntilExpiry() int {
	if s.IsExpired() {
		return 0
	}
	return int(s.EndDate.Sub(time.Now()).Hours() / 24)
}

// GetMonthlyPrice returns monthly price for the subscription
func (s *Subscription) GetMonthlyPrice() float64 {
	if s.PlanType == "monthly" {
		return s.Price
	}
	return s.Price / 12 // Convert yearly to monthly
}

// CanAddVehicle checks if company can add more vehicles
func (s *Subscription) CanAddVehicle(currentVehicleCount int) bool {
	return currentVehicleCount < s.MaxVehicles
}

// CanAddDriver checks if company can add more drivers
func (s *Subscription) CanAddDriver(currentDriverCount int) bool {
	return currentDriverCount < s.MaxDrivers
}

// CanAddUser checks if company can add more users
func (s *Subscription) CanAddUser(currentUserCount int) bool {
	return currentUserCount < s.MaxUsers
}

// HasFeature checks if subscription includes a specific feature
func (s *Subscription) HasFeature(feature string) bool {
	if s.Features == nil {
		return false
	}
	value, exists := s.Features[feature]
	if !exists {
		return false
	}
	
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	return false
}

// Payment methods
// IsCompleted checks if payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == "completed"
}

// IsPending checks if payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == "pending"
}

// IsFailed checks if payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == "failed"
}

// IsExpired checks if payment is expired
func (p *Payment) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}

// CompletePayment marks payment as completed
func (p *Payment) CompletePayment(externalID string) {
	p.Status = "completed"
	now := time.Now()
	p.CompletedAt = &now
	p.ExternalID = externalID
}

// FailPayment marks payment as failed
func (p *Payment) FailPayment(reason string) {
	p.Status = "failed"
	if p.GatewayResponse == nil {
		p.GatewayResponse = make(JSON)
	}
	p.GatewayResponse["failure_reason"] = reason
}

// GetQRISData returns QRIS payment data
func (p *Payment) GetQRISData() map[string]interface{} {
	if p.QRISData == nil {
		return make(map[string]interface{})
	}
	return p.QRISData
}

// SetQRISData sets QRIS payment data
func (p *Payment) SetQRISData(data map[string]interface{}) {
	p.QRISData = data
}

// GetBankTransferData returns bank transfer data
func (p *Payment) GetBankTransferData() map[string]interface{} {
	if p.BankTransfer == nil {
		return make(map[string]interface{})
	}
	return p.BankTransfer
}

// SetBankTransferData sets bank transfer data
func (p *Payment) SetBankTransferData(data map[string]interface{}) {
	p.BankTransfer = data
}

// GetEWalletData returns e-wallet payment data
func (p *Payment) GetEWalletData() map[string]interface{} {
	if p.EWalletData == nil {
		return make(map[string]interface{})
	}
	return p.EWalletData
}

// SetEWalletData sets e-wallet payment data
func (p *Payment) SetEWalletData(data map[string]interface{}) {
	p.EWalletData = data
}

// Invoice methods
// IsPaid checks if invoice is paid
func (i *Invoice) IsPaid() bool {
	return i.Status == "paid" || i.BalanceAmount <= 0
}

// IsOverdue checks if invoice is overdue
func (i *Invoice) IsOverdue() bool {
	return !i.IsPaid() && time.Now().After(i.DueDate)
}

// DaysOverdue returns days overdue
func (i *Invoice) DaysOverdue() int {
	if !i.IsOverdue() {
		return 0
	}
	return int(time.Now().Sub(i.DueDate).Hours() / 24)
}

// MarkAsPaid marks invoice as paid
func (i *Invoice) MarkAsPaid(paymentAmount float64, paymentMethod string) {
	i.PaidAmount += paymentAmount
	i.BalanceAmount = i.TotalAmount - i.PaidAmount
	
	if i.BalanceAmount <= 0 {
		i.Status = "paid"
	}
	
	i.PaymentMethod = paymentMethod
}

// AddPayment adds a payment to the invoice
func (i *Invoice) AddPayment(payment *Payment) {
	i.PaymentID = &payment.ID
	i.MarkAsPaid(payment.TotalAmount, payment.PaymentMethod)
}

// GetFormattedAmount returns formatted amount string
func (i *Invoice) GetFormattedAmount() string {
	return formatCurrency(i.TotalAmount, i.Currency)
}

// GetFormattedBalance returns formatted balance string
func (i *Invoice) GetFormattedBalance() string {
	return formatCurrency(i.BalanceAmount, i.Currency)
}

// Helper function to format currency
func formatCurrency(amount float64, currency string) string {
	switch currency {
	case "IDR":
		return "Rp " + formatNumber(amount)
	default:
		return currency + " " + formatNumber(amount)
	}
}

// Helper function to format numbers with thousand separators
func formatNumber(num float64) string {
	// Simple formatting - can be enhanced with proper locale formatting
	return fmt.Sprintf("%.2f", num)
}
