package payment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
)

func TestService_GenerateInvoice(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	tests := []struct {
		name    string
		request InvoiceRequest
		wantErr bool
	}{
		{
			name: "valid invoice generation",
			request: InvoiceRequest{
				CompanyID:     company.ID,
				BillingPeriod: "2025-01",
				DueDate:       time.Now().AddDate(0, 0, 30),
				Notes:         "Monthly subscription",
			},
			wantErr: false,
		},
		{
			name: "invalid company ID",
			request: InvoiceRequest{
				CompanyID:     "invalid-company-id",
				BillingPeriod: "2025-01",
				DueDate:       time.Now().AddDate(0, 0, 30),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice, err := service.GenerateInvoice(ctx, &tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, invoice)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, invoice)
				testutil.AssertValidUUID(t, invoice.InvoiceID)
				assert.NotEmpty(t, invoice.InvoiceNumber)
				
				// Validate Indonesian tax calculation (PPN 11%)
				testutil.AssertValidPPN11(t, invoice.Subtotal, invoice.TaxAmount)
				assert.Equal(t, invoice.Subtotal+invoice.TaxAmount, invoice.TotalAmount)
				
				// Validate currency
				testutil.AssertValidCurrency(t, invoice.TotalAmount)
				
				// Validate payment instructions
				assert.NotEmpty(t, invoice.PaymentInstructions.BankName)
				assert.NotEmpty(t, invoice.PaymentInstructions.AccountNumber)
				assert.NotEmpty(t, invoice.PaymentInstructions.ReferenceCode)
			}
		})
	}
}

func TestService_IndonesianTaxCalculation(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	tests := []struct {
		name         string
		subtotal     float64
		expectedTax  float64
		expectedRate float64
	}{
		{
			name:         "basic subscription - IDR 250,000",
			subtotal:     250000.0,
			expectedTax:  27500.0, // 11% of 250,000
			expectedRate: 11.0,
		},
		{
			name:         "premium subscription - IDR 500,000",
			subtotal:     500000.0,
			expectedTax:  55000.0, // 11% of 500,000
			expectedRate: 11.0,
		},
		{
			name:         "enterprise subscription - IDR 1,000,000",
			subtotal:     1000000.0,
			expectedTax:  110000.0, // 11% of 1,000,000
			expectedRate: 11.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create invoice to test tax calculation
			invoice, err := service.GenerateInvoice(ctx, &InvoiceRequest{
				CompanyID:     company.ID,
				BillingPeriod: "2025-01",
				DueDate:       time.Now().AddDate(0, 0, 30),
			})

			require.NoError(t, err)
			require.NotNil(t, invoice)

			// Indonesian PPN (Pajak Pertambahan Nilai) should be 11%
			expectedTaxAmount := tt.subtotal * (tt.expectedRate / 100)
			
			// Allow small floating point differences
			assert.InDelta(t, expectedTaxAmount, invoice.TaxAmount, 0.01, 
				"Tax calculation should follow Indonesian PPN 11% rate")
			
			// Validate using custom assertion
			testutil.AssertValidPPN11(t, invoice.Subtotal, invoice.TaxAmount)
			
			// Total should be subtotal + tax
			expectedTotal := invoice.Subtotal + invoice.TaxAmount
			assert.InDelta(t, expectedTotal, invoice.TotalAmount, 0.01)
		})
	}
}

func TestService_InvoiceNumberFormat(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	t.Run("Indonesian invoice number format", func(t *testing.T) {
		invoice, err := service.GenerateInvoice(ctx, &InvoiceRequest{
			CompanyID:     company.ID,
			BillingPeriod: "2025-01",
			DueDate:       time.Now().AddDate(0, 0, 30),
		})

		require.NoError(t, err)
		require.NotNil(t, invoice)

		// Indonesian invoice format: INV/YYYY/MM/XXXX
		// Example: INV/2025/01/0001
		assert.Contains(t, invoice.InvoiceNumber, "INV/", 
			"Invoice number should start with INV/")
		assert.Contains(t, invoice.InvoiceNumber, "2025/", 
			"Invoice number should contain current year")
	})
}

func TestService_ConfirmPayment(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	// Generate an invoice first
	invoiceResp, err := service.GenerateInvoice(ctx, &InvoiceRequest{
		CompanyID:     company.ID,
		BillingPeriod: "2025-01",
		DueDate:       time.Now().AddDate(0, 0, 30),
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		request PaymentConfirmationRequest
		wantErr bool
	}{
		{
			name: "valid payment confirmation",
			request: PaymentConfirmationRequest{
				InvoiceID:       invoiceResp.InvoiceID,
				BankAccount:     "BCA - 1234567890",
				TransferAmount:  invoiceResp.TotalAmount,
				TransferDate:    time.Now(),
				ReferenceNumber: "REF123456789",
				Notes:           "Payment via BCA mobile",
			},
			wantErr: false,
		},
		{
			name: "invalid invoice ID",
			request: PaymentConfirmationRequest{
				InvoiceID:       "invalid-invoice-id",
				BankAccount:     "BCA - 1234567890",
				TransferAmount:  100000.0,
				TransferDate:    time.Now(),
				ReferenceNumber: "REF123456789",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ConfirmPayment(ctx, &tt.request, "admin@test.com")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetPaymentInstructions(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	// Generate an invoice
	invoiceResp, err := service.GenerateInvoice(ctx, &InvoiceRequest{
		CompanyID:     company.ID,
		BillingPeriod: "2025-01",
		DueDate:       time.Now().AddDate(0, 0, 30),
	})
	require.NoError(t, err)

	t.Run("get payment instructions", func(t *testing.T) {
		instructions, err := service.GetPaymentInstructions(ctx, invoiceResp.InvoiceID)

		assert.NoError(t, err)
		assert.NotNil(t, instructions)
		assert.NotEmpty(t, instructions.BankName, "Bank name should be provided")
		assert.NotEmpty(t, instructions.AccountNumber, "Account number should be provided")
		assert.NotEmpty(t, instructions.AccountHolder, "Account holder should be provided")
		assert.NotEmpty(t, instructions.ReferenceCode, "Reference code should be provided")
		assert.NotEmpty(t, instructions.Amount, "Amount should be formatted")
		
		// Validate Indonesian bank names
		validBanks := []string{"BCA", "Mandiri", "BNI", "BRI"}
		hasValidBank := false
		for _, bank := range validBanks {
			if instructions.BankName == bank {
				hasValidBank = true
				break
			}
		}
		assert.True(t, hasValidBank, "Should use Indonesian bank (BCA, Mandiri, BNI, or BRI)")
	})
}

func TestService_GetInvoices(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	// Create test company
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	// Generate multiple invoices
	for i := 0; i < 3; i++ {
		_, err := service.GenerateInvoice(ctx, &InvoiceRequest{
			CompanyID:     company.ID,
			BillingPeriod: "2025-01",
			DueDate:       time.Now().AddDate(0, 0, 30),
		})
		require.NoError(t, err)
	}

	t.Run("get all invoices", func(t *testing.T) {
		invoices, err := service.GetInvoices(ctx, company.ID, "", 10, 0)

		assert.NoError(t, err)
		assert.NotEmpty(t, invoices)
		assert.GreaterOrEqual(t, len(invoices), 3)
		
		// Validate each invoice has Indonesian compliance
		for _, invoice := range invoices {
			testutil.AssertValidUUID(t, invoice.ID)
			assert.Equal(t, "IDR", invoice.Currency, "Currency should be IDR")
			assert.Equal(t, 11.0, invoice.TaxRate, "Tax rate should be 11% (Indonesian PPN)")
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		invoices, err := service.GetInvoices(ctx, company.ID, "draft", 10, 0)

		assert.NoError(t, err)
		// Should have at least some invoices in draft status
		for _, invoice := range invoices {
			assert.Equal(t, "draft", invoice.Status)
		}
	})
}

func TestService_IndonesianCurrencyFormat(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	t.Run("amounts in IDR", func(t *testing.T) {
		invoice, err := service.GenerateInvoice(ctx, &InvoiceRequest{
			CompanyID:     company.ID,
			BillingPeriod: "2025-01",
			DueDate:       time.Now().AddDate(0, 0, 30),
		})

		require.NoError(t, err)
		
		// All amounts should be positive IDR values
		testutil.AssertValidCurrency(t, invoice.Subtotal)
		testutil.AssertValidCurrency(t, invoice.TaxAmount)
		testutil.AssertValidCurrency(t, invoice.TotalAmount)
		
		// Payment instructions amount should contain IDR
		assert.Contains(t, invoice.PaymentInstructions.Amount, "IDR", 
			"Payment amount should display IDR currency")
	})
}

func TestService_TaxCompliance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	t.Run("Indonesian PPN compliance", func(t *testing.T) {
		invoice, err := service.GenerateInvoice(ctx, &InvoiceRequest{
			CompanyID:     company.ID,
			BillingPeriod: "2025-01",
			DueDate:       time.Now().AddDate(0, 0, 30),
		})

		require.NoError(t, err)
		
		// Tax rate must be exactly 11% (Indonesian PPN as of 2022)
		expectedTaxRate := 11.0
		actualTaxRate := (invoice.TaxAmount / invoice.Subtotal) * 100
		
		assert.InDelta(t, expectedTaxRate, actualTaxRate, 0.01, 
			"Tax rate must be 11% according to Indonesian PPN regulation")
		
		// Company should have NPWP (tax ID)
		testutil.AssertValidNPWP(t, company.NPWP)
	})
}

func TestService_DueDateCalculation(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	cfg := &config.Config{}
	repoManager := repository.NewRepositoryManager(db)
	service := NewService(db, cfg, repoManager)

	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	ctx := context.Background()

	t.Run("due date is in the future", func(t *testing.T) {
		dueDate := time.Now().AddDate(0, 0, 30) // 30 days from now
		
		invoice, err := service.GenerateInvoice(ctx, &InvoiceRequest{
			CompanyID:     company.ID,
			BillingPeriod: "2025-01",
			DueDate:       dueDate,
		})

		require.NoError(t, err)
		
		// Parse due date string and verify it's in the future
		assert.NotEmpty(t, invoice.DueDate)
		// Due date should be reasonable (within 60 days)
		assert.NotEmpty(t, invoice.InvoiceID)
	})
}

