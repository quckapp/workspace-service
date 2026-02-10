package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/quckapp/workspace-service/internal/models"
	"github.com/quckapp/workspace-service/internal/repository"
	"github.com/sirupsen/logrus"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrPaymentMethodNotFound = errors.New("payment method not found")
	ErrCannotDowngrade      = errors.New("cannot downgrade with current usage")
	ErrInsufficientSeats    = errors.New("cannot remove seats below current member count")
	ErrAlreadyOnPlan        = errors.New("workspace is already on this plan")
)

type BillingService struct {
	billingRepo *repository.BillingRepository
	memberRepo  *repository.MemberRepository
	logger      *logrus.Logger
}

func NewBillingService(billingRepo *repository.BillingRepository, memberRepo *repository.MemberRepository, logger *logrus.Logger) *BillingService {
	return &BillingService{billingRepo: billingRepo, memberRepo: memberRepo, logger: logger}
}

func (s *BillingService) GetBillingOverview(ctx context.Context, workspaceID uuid.UUID) (*models.BillingOverview, error) {
	plan, err := s.billingRepo.GetPlan(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	overview := &models.BillingOverview{
		Plan: plan,
	}

	if plan != nil {
		overview.NextBillingDate = &plan.CurrentPeriodEnd
		overview.EstimatedCost = plan.SeatCount * plan.PricePerSeat

		pm, _ := s.billingRepo.GetDefaultPaymentMethod(ctx, workspaceID)
		overview.PaymentMethod = pm

		invoices, _ := s.billingRepo.ListInvoices(ctx, workspaceID, 1, 0)
		if len(invoices) > 0 {
			overview.CurrentInvoice = invoices[0]
		}
	}

	return overview, nil
}

func (s *BillingService) GetPlan(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspacePlan, error) {
	plan, err := s.billingRepo.GetPlan(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}
	return plan, nil
}

func (s *BillingService) ChangePlan(ctx context.Context, workspaceID, userID uuid.UUID, req *models.ChangePlanRequest) (*models.WorkspacePlan, error) {
	plan, _ := s.billingRepo.GetPlan(ctx, workspaceID)

	now := time.Now()
	features := s.GetPlanFeatures(req.PlanType)

	if plan == nil {
		plan = &models.WorkspacePlan{
			ID:                 uuid.New(),
			WorkspaceID:        workspaceID,
			PlanType:           req.PlanType,
			Status:             "active",
			BillingCycle:       req.BillingCycle,
			SeatCount:          1,
			SeatLimit:          features.MaxMembers,
			StorageLimitMB:     features.MaxStorageMB,
			PricePerSeat:       features.PricePerSeat,
			Currency:           "usd",
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now.AddDate(0, 1, 0),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if req.BillingCycle == "annual" {
			plan.CurrentPeriodEnd = now.AddDate(1, 0, 0)
		}
		if err := s.billingRepo.CreatePlan(ctx, plan); err != nil {
			return nil, err
		}
	} else {
		if plan.PlanType == req.PlanType {
			return nil, ErrAlreadyOnPlan
		}
		plan.PlanType = req.PlanType
		plan.BillingCycle = req.BillingCycle
		plan.SeatLimit = features.MaxMembers
		plan.StorageLimitMB = features.MaxStorageMB
		plan.PricePerSeat = features.PricePerSeat
		plan.Status = "active"
		plan.CanceledAt = nil
		if err := s.billingRepo.UpdatePlan(ctx, plan); err != nil {
			return nil, err
		}
	}

	event := &models.BillingEvent{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		EventType:   "plan_changed",
		Description: fmt.Sprintf("Plan changed to %s (%s)", req.PlanType, req.BillingCycle),
		ActorID:     userID,
		CreatedAt:   now,
	}
	s.billingRepo.CreateEvent(ctx, event)

	return plan, nil
}

func (s *BillingService) CancelPlan(ctx context.Context, workspaceID, userID uuid.UUID) error {
	plan, err := s.billingRepo.GetPlan(ctx, workspaceID)
	if err != nil || plan == nil {
		return ErrPlanNotFound
	}

	now := time.Now()
	plan.Status = "canceled"
	plan.CanceledAt = &now

	if err := s.billingRepo.UpdatePlan(ctx, plan); err != nil {
		return err
	}

	event := &models.BillingEvent{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		EventType:   "plan_canceled",
		Description: "Plan canceled",
		ActorID:     userID,
		CreatedAt:   now,
	}
	s.billingRepo.CreateEvent(ctx, event)

	return nil
}

func (s *BillingService) AddSeats(ctx context.Context, workspaceID, userID uuid.UUID, req *models.AddSeatsRequest) (*models.WorkspacePlan, error) {
	plan, err := s.billingRepo.GetPlan(ctx, workspaceID)
	if err != nil || plan == nil {
		return nil, ErrPlanNotFound
	}

	plan.SeatCount += req.Count
	if err := s.billingRepo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}

	event := &models.BillingEvent{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		EventType:   "seat_added",
		Description: fmt.Sprintf("Added %d seats", req.Count),
		ActorID:     userID,
		CreatedAt:   time.Now(),
	}
	s.billingRepo.CreateEvent(ctx, event)

	return plan, nil
}

func (s *BillingService) RemoveSeats(ctx context.Context, workspaceID, userID uuid.UUID, req *models.RemoveSeatsRequest) (*models.WorkspacePlan, error) {
	plan, err := s.billingRepo.GetPlan(ctx, workspaceID)
	if err != nil || plan == nil {
		return nil, ErrPlanNotFound
	}

	newCount := plan.SeatCount - req.Count
	if newCount < 1 {
		return nil, ErrInsufficientSeats
	}

	plan.SeatCount = newCount
	if err := s.billingRepo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}

	event := &models.BillingEvent{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		EventType:   "seat_removed",
		Description: fmt.Sprintf("Removed %d seats", req.Count),
		ActorID:     userID,
		CreatedAt:   time.Now(),
	}
	s.billingRepo.CreateEvent(ctx, event)

	return plan, nil
}

func (s *BillingService) ListInvoices(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.BillingInvoice, error) {
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.billingRepo.ListInvoices(ctx, workspaceID, perPage, offset)
}

func (s *BillingService) GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*models.BillingInvoice, error) {
	invoice, err := s.billingRepo.GetInvoice(ctx, invoiceID)
	if err != nil || invoice == nil {
		return nil, ErrInvoiceNotFound
	}
	return invoice, nil
}

func (s *BillingService) ListPaymentMethods(ctx context.Context, workspaceID uuid.UUID) ([]*models.PaymentMethod, error) {
	return s.billingRepo.ListPaymentMethods(ctx, workspaceID)
}

func (s *BillingService) AddPaymentMethod(ctx context.Context, workspaceID, userID uuid.UUID, req *models.AddPaymentMethodRequest) (*models.PaymentMethod, error) {
	now := time.Now()
	pm := &models.PaymentMethod{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Type:        req.Type,
		Last4:       "0000",
		ExpMonth:    12,
		ExpYear:     now.Year() + 3,
		IsDefault:   false,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	existing, _ := s.billingRepo.ListPaymentMethods(ctx, workspaceID)
	if len(existing) == 0 {
		pm.IsDefault = true
	}

	if err := s.billingRepo.CreatePaymentMethod(ctx, pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (s *BillingService) SetDefaultPaymentMethod(ctx context.Context, workspaceID, methodID uuid.UUID) error {
	pm, err := s.billingRepo.GetPaymentMethod(ctx, methodID)
	if err != nil || pm == nil {
		return ErrPaymentMethodNotFound
	}
	return s.billingRepo.SetDefaultPaymentMethod(ctx, workspaceID, methodID)
}

func (s *BillingService) DeletePaymentMethod(ctx context.Context, methodID uuid.UUID) error {
	pm, err := s.billingRepo.GetPaymentMethod(ctx, methodID)
	if err != nil || pm == nil {
		return ErrPaymentMethodNotFound
	}
	return s.billingRepo.DeletePaymentMethod(ctx, methodID)
}

func (s *BillingService) ListBillingEvents(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.BillingEvent, error) {
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.billingRepo.ListEvents(ctx, workspaceID, perPage, offset)
}

func (s *BillingService) GetAvailablePlans() []models.PlanFeatures {
	return []models.PlanFeatures{
		s.GetPlanFeatures("free"),
		s.GetPlanFeatures("starter"),
		s.GetPlanFeatures("pro"),
		s.GetPlanFeatures("business"),
		s.GetPlanFeatures("enterprise"),
	}
}

func (s *BillingService) GetPlanFeatures(planType string) models.PlanFeatures {
	switch planType {
	case "starter":
		return models.PlanFeatures{PlanType: "starter", MaxMembers: 25, MaxChannels: 100, MaxStorageMB: 10240, MaxIntegrations: 5, CustomEmoji: true, AdvancedSecurity: false, AuditLogs: false, Compliance: false, SSO: false, GuestAccess: true, PricePerSeat: 500}
	case "pro":
		return models.PlanFeatures{PlanType: "pro", MaxMembers: 100, MaxChannels: 500, MaxStorageMB: 51200, MaxIntegrations: 20, CustomEmoji: true, AdvancedSecurity: true, AuditLogs: true, Compliance: false, SSO: false, GuestAccess: true, PricePerSeat: 1000}
	case "business":
		return models.PlanFeatures{PlanType: "business", MaxMembers: 500, MaxChannels: 2000, MaxStorageMB: 204800, MaxIntegrations: 50, CustomEmoji: true, AdvancedSecurity: true, AuditLogs: true, Compliance: true, SSO: true, GuestAccess: true, PricePerSeat: 1500}
	case "enterprise":
		return models.PlanFeatures{PlanType: "enterprise", MaxMembers: 10000, MaxChannels: 10000, MaxStorageMB: 1048576, MaxIntegrations: 100, CustomEmoji: true, AdvancedSecurity: true, AuditLogs: true, Compliance: true, SSO: true, GuestAccess: true, PricePerSeat: 2500}
	default: // free
		return models.PlanFeatures{PlanType: "free", MaxMembers: 10, MaxChannels: 20, MaxStorageMB: 5120, MaxIntegrations: 2, CustomEmoji: false, AdvancedSecurity: false, AuditLogs: false, Compliance: false, SSO: false, GuestAccess: false, PricePerSeat: 0}
	}
}
