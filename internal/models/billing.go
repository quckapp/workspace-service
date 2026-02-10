package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Billing & Plans ──

type WorkspacePlan struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WorkspaceID     uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	PlanType        string     `json:"plan_type" db:"plan_type"` // free, starter, pro, business, enterprise
	Status          string     `json:"status" db:"status"`       // active, trialing, past_due, canceled, paused
	BillingCycle    string     `json:"billing_cycle" db:"billing_cycle"` // monthly, annual
	SeatCount       int        `json:"seat_count" db:"seat_count"`
	SeatLimit       int        `json:"seat_limit" db:"seat_limit"`
	StorageLimitMB  int64      `json:"storage_limit_mb" db:"storage_limit_mb"`
	StorageUsedMB   int64      `json:"storage_used_mb" db:"storage_used_mb"`
	PricePerSeat    int        `json:"price_per_seat" db:"price_per_seat"` // cents
	Currency        string     `json:"currency" db:"currency"`
	TrialEndsAt     *time.Time `json:"trial_ends_at" db:"trial_ends_at"`
	CurrentPeriodStart time.Time `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" db:"current_period_end"`
	CanceledAt      *time.Time `json:"canceled_at" db:"canceled_at"`
	ExternalID      *string    `json:"external_id" db:"external_id"` // Stripe subscription ID
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type BillingInvoice struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	WorkspaceID   uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	InvoiceNumber string     `json:"invoice_number" db:"invoice_number"`
	Amount        int        `json:"amount" db:"amount"` // cents
	Currency      string     `json:"currency" db:"currency"`
	Status        string     `json:"status" db:"status"` // draft, open, paid, void, uncollectible
	Description   *string    `json:"description" db:"description"`
	PeriodStart   time.Time  `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time  `json:"period_end" db:"period_end"`
	PaidAt        *time.Time `json:"paid_at" db:"paid_at"`
	DueDate       *time.Time `json:"due_date" db:"due_date"`
	ExternalID    *string    `json:"external_id" db:"external_id"` // Stripe invoice ID
	PDFURL        *string    `json:"pdf_url" db:"pdf_url"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type PaymentMethod struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Type        string     `json:"type" db:"type"` // card, bank_account
	Brand       *string    `json:"brand" db:"brand"` // visa, mastercard, amex
	Last4       string     `json:"last4" db:"last4"`
	ExpMonth    int        `json:"exp_month" db:"exp_month"`
	ExpYear     int        `json:"exp_year" db:"exp_year"`
	IsDefault   bool       `json:"is_default" db:"is_default"`
	ExternalID  *string    `json:"external_id" db:"external_id"` // Stripe payment method ID
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type BillingEvent struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	EventType   string    `json:"event_type" db:"event_type"` // plan_changed, payment_success, payment_failed, seat_added, seat_removed
	Description string    `json:"description" db:"description"`
	Metadata    JSON      `json:"metadata" db:"metadata"`
	ActorID     uuid.UUID `json:"actor_id" db:"actor_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ChangePlanRequest struct {
	PlanType     string `json:"plan_type" binding:"required,oneof=free starter pro business enterprise"`
	BillingCycle string `json:"billing_cycle" binding:"required,oneof=monthly annual"`
}

type AddSeatsRequest struct {
	Count int `json:"count" binding:"required,min=1,max=500"`
}

type RemoveSeatsRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}

type AddPaymentMethodRequest struct {
	Type     string `json:"type" binding:"required,oneof=card bank_account"`
	Token    string `json:"token" binding:"required"` // payment token from client-side
}

type BillingOverview struct {
	Plan           *WorkspacePlan   `json:"plan"`
	CurrentInvoice *BillingInvoice  `json:"current_invoice,omitempty"`
	PaymentMethod  *PaymentMethod   `json:"default_payment_method,omitempty"`
	NextBillingDate *time.Time      `json:"next_billing_date,omitempty"`
	EstimatedCost  int              `json:"estimated_cost"` // cents
}

type PlanFeatures struct {
	PlanType       string `json:"plan_type"`
	MaxMembers     int    `json:"max_members"`
	MaxChannels    int    `json:"max_channels"`
	MaxStorageMB   int64  `json:"max_storage_mb"`
	MaxIntegrations int   `json:"max_integrations"`
	CustomEmoji    bool   `json:"custom_emoji"`
	AdvancedSecurity bool `json:"advanced_security"`
	AuditLogs      bool   `json:"audit_logs"`
	Compliance     bool   `json:"compliance"`
	SSO            bool   `json:"sso"`
	GuestAccess    bool   `json:"guest_access"`
	PricePerSeat   int    `json:"price_per_seat"` // cents/month
}

type UsageReport struct {
	WorkspaceID    uuid.UUID         `json:"workspace_id"`
	Period         string            `json:"period"`
	ActiveMembers  int               `json:"active_members"`
	TotalMessages  int               `json:"total_messages"`
	StorageUsedMB  int64             `json:"storage_used_mb"`
	APICallCount   int               `json:"api_call_count"`
	IntegrationCount int             `json:"integration_count"`
	DailyUsage     []DailyUsageEntry `json:"daily_usage"`
}

type DailyUsageEntry struct {
	Date           string `json:"date" db:"date"`
	ActiveUsers    int    `json:"active_users" db:"active_users"`
	Messages       int    `json:"messages" db:"messages"`
	StorageDeltaMB int64  `json:"storage_delta_mb" db:"storage_delta_mb"`
}
