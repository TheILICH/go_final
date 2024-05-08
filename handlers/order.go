package handlers

import (
	"github.com/gin-gonic/gin"
	"go_final/models"
	"go_final/repositories"
	"net/http"
	"strconv"
)

type OrderHandler interface {
	OrderProducts(*gin.Context)
	UpdateOrder(*gin.Context)
	UpdateOrderStatus(ctx *gin.Context)
	DeleteOrder(*gin.Context)
	DeleteOrderItem(*gin.Context)
	GetOrderItems(*gin.Context)
	GetOrders(*gin.Context)
	GetOrderByID(*gin.Context)
}

type orderHandler struct {
	repo repositories.OrderRepository
}

func NewOrderHandler() OrderHandler {
	return &orderHandler{
		repo: repositories.NewOrderRepository(),
	}
}

func (h *orderHandler) GetOrderByID(ctx *gin.Context) {
	orderIDStr := ctx.Param("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	order, err := h.repo.GetOrderByID(uint(orderID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

func (h *orderHandler) GetOrders(ctx *gin.Context) {
	userID := ctx.GetFloat64("userID")
	orders, err := h.repo.GetOrders(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

func (h *orderHandler) GetOrderItems(ctx *gin.Context) {
	orderIDStr := ctx.Param("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	order, err := h.repo.GetOrderItems(uint(orderID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

func (h *orderHandler) OrderProducts(ctx *gin.Context) {
	var input models.CartRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := ctx.GetFloat64("userID")
	if err := h.repo.OrderProducts(uint(userID), input.CartItems); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, input)
}

func (h *orderHandler) UpdateOrder(ctx *gin.Context) {
	var input models.CartRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := ctx.GetFloat64("userID")
	if err := h.repo.UpdateOrder(uint(userID), input.CartItems); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, input)
}

func (h *orderHandler) DeleteOrder(ctx *gin.Context) {
	userID := ctx.GetFloat64("userID")
	if err := h.repo.DeleteOrder(uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete order"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *orderHandler) DeleteOrderItem(ctx *gin.Context) {
	userID := ctx.GetFloat64("userID")

	orderItemIDStr := ctx.Param("order_item_id")
	orderItemID, err := strconv.Atoi(orderItemIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order item ID"})
		return
	}

	if err = h.repo.DeleteOrderItem(uint(userID), uint(orderItemID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete order item"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *orderHandler) UpdateOrderStatus(ctx *gin.Context) {
	userID := ctx.GetFloat64("userID")
	userStatus := ctx.GetString("role")
	orderIDStr := ctx.Param("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	newStatus := ctx.Param("status")

	order, err := h.repo.GetOrderByID(uint(orderID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.UserID != uint(userID) && !isStatusTransitionAllowed(order.OrderStatus, models.OrderStatus(newStatus), userStatus) {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "Status transition not allowed" + strconv.FormatBool(order.UserID == uint(userID)) + strconv.FormatBool(isStatusTransitionAllowed(order.OrderStatus, models.OrderStatus(newStatus), userStatus))},
		)
		return
	}

	if err = h.repo.UpdateOrderStatus(uint(orderID), newStatus); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

func isStatusTransitionAllowed(currentStatus, newStatus models.OrderStatus, userRole string) bool {
	switch userRole {
	case "admin":
		return isAdminTransitionValid(currentStatus, newStatus)
	case "customer":
		return isCustomerTransitionValid(currentStatus, newStatus)
	}
	return false
}

func isAdminTransitionValid(current, new models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus]models.OrderStatus{
		models.PENDING:  models.ACCEPTED,
		models.ACCEPTED: models.READY,
		models.READY:    models.OUT,
		models.OUT:      models.DELIVERED,
	}
	nextStatus, ok := validTransitions[current]
	return ok && new == nextStatus
}

func isCustomerTransitionValid(current, new models.OrderStatus) bool {
	if new == models.CANCELED {
		return true
	}
	if current == models.DELIVERED && new == models.CONFIRMED {
		return true
	}
	return false
}
