package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quckapp/workspace-service/internal/models"
	"github.com/quckapp/workspace-service/internal/service"
	"github.com/sirupsen/logrus"
)

type BillingHandler struct {
	service *service.BillingService
	logger  *logrus.Logger
}

func NewBillingHandler(svc *service.BillingService, logger *logrus.Logger) *BillingHandler {
	return &BillingHandler{service: svc, logger: logger}
}

func (h *BillingHandler) GetBillingOverview(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	overview, err := h.service.GetBillingOverview(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get billing overview"})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *BillingHandler) GetPlan(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	plan, err := h.service.GetPlan(c.Request.Context(), workspaceID)
	if err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *BillingHandler) ChangePlan(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.ChangePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.service.ChangePlan(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *BillingHandler) CancelPlan(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	if err := h.service.CancelPlan(c.Request.Context(), workspaceID, userID); err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Plan canceled"})
}

func (h *BillingHandler) AddSeats(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.AddSeatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.service.AddSeats(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *BillingHandler) RemoveSeats(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.RemoveSeatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.service.RemoveSeats(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	invoices, err := h.service.ListInvoices(c.Request.Context(), workspaceID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list invoices"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	invoiceID, _ := uuid.Parse(c.Param("invoiceId"))
	invoice, err := h.service.GetInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) ListPaymentMethods(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	methods, err := h.service.ListPaymentMethods(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list payment methods"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment_methods": methods})
}

func (h *BillingHandler) AddPaymentMethod(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pm, err := h.service.AddPaymentMethod(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add payment method"})
		return
	}
	c.JSON(http.StatusCreated, pm)
}

func (h *BillingHandler) SetDefaultPaymentMethod(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	methodID, _ := uuid.Parse(c.Param("methodId"))
	if err := h.service.SetDefaultPaymentMethod(c.Request.Context(), workspaceID, methodID); err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Default payment method updated"})
}

func (h *BillingHandler) DeletePaymentMethod(c *gin.Context) {
	methodID, _ := uuid.Parse(c.Param("methodId"))
	if err := h.service.DeletePaymentMethod(c.Request.Context(), methodID); err != nil {
		billingHandleError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *BillingHandler) ListBillingEvents(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	events, err := h.service.ListBillingEvents(c.Request.Context(), workspaceID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list billing events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (h *BillingHandler) GetAvailablePlans(c *gin.Context) {
	plans := h.service.GetAvailablePlans()
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *BillingHandler) GetPlanFeatures(c *gin.Context) {
	planType := c.Param("planType")
	features := h.service.GetPlanFeatures(planType)
	c.JSON(http.StatusOK, features)
}

func billingHandleError(c *gin.Context, err error) {
	switch err {
	case service.ErrPlanNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
	case service.ErrInvoiceNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
	case service.ErrPaymentMethodNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
	case service.ErrCannotDowngrade:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot downgrade with current usage"})
	case service.ErrInsufficientSeats:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove seats below current member count"})
	case service.ErrAlreadyOnPlan:
		c.JSON(http.StatusConflict, gin.H{"error": "Already on this plan"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
