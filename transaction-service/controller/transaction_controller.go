package controller

import (
	"fmt"
	"micro-warehouse/transaction-service/controller/request"
	"micro-warehouse/transaction-service/controller/response"
	"micro-warehouse/transaction-service/model"
	"micro-warehouse/transaction-service/pkg/conv"
	"micro-warehouse/transaction-service/pkg/midtrans"
	"micro-warehouse/transaction-service/pkg/pagination"
	"micro-warehouse/transaction-service/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type TransactionControllerInterface interface {
	CreateTransaction(ctx *fiber.Ctx) error
	GetTransactions(c *fiber.Ctx) error
	GetTransactionByID(c *fiber.Ctx) error
	MidtransCallback(c *fiber.Ctx) error

	GetManagerDashboard(c *fiber.Ctx) error
	GetDashboardByMerchant(c *fiber.Ctx) error
}

type transactionController struct {
	transactionUsecase usecase.TransactionUsecaseInterface
	midtransService    midtrans.MidtransServiceInterface
}

// CreateTransaction implements TransactionControllerInterface.
func (t *transactionController) CreateTransaction(ctx *fiber.Ctx) error {
	var req request.CreateTransactionWithProductsRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Errorf("[TransactionController] CreateTransaction - 1: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	orderID := fmt.Sprintf("ORDER_%d_%d", time.Now().Unix(), req.MerchantID)

	var subtotal int64
	var items []midtrans.TransactionItem

	for _, product := range req.Products {
		productSubtotal := int64(product.Price) * product.Quantity
		subtotal += productSubtotal

		items = append(items, midtrans.TransactionItem{
			ID:       fmt.Sprintf("%d", product.ProductID),
			Price:    product.Price,
			Quantity: product.Quantity,
			Name:     fmt.Sprintf("Product %d", product.ProductID),
		})
	}

	taxtotal := int64(float64(subtotal) * 0.11)
	grandtotal := subtotal + taxtotal

	transaction := model.Transaction{
		Name:          req.Name,
		Phone:         req.Phone,
		Email:         req.Email,
		Address:       req.Address,
		SubTotal:      subtotal,
		TaxTotal:      taxtotal,
		GrandTotal:    grandtotal,
		MerchantID:    req.MerchantID,
		Notes:         req.Notes,
		Currency:      "IDR",
		OrderID:       orderID,
		PaymentStatus: model.PaymentStatusPending,
		PaymentMethod: model.PaymentMethodQRIS,
	}

	for _, product := range req.Products {
		transaction.TransactionProducts = append(transaction.TransactionProducts, model.TransactionProduct{
			ProductID: product.ProductID,
			Quantity:  product.Quantity,
			Price:     product.Price,
			SubTotal:  product.Price * product.Quantity,
		})
	}

	idTransaction, err := t.transactionUsecase.CreateTransaction(ctx.Context(), transaction)
	if err != nil {
		log.Errorf("[TransactionController] CreateTransaction - 2: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create transaction",
		})
	}

	midtransReq := midtrans.CreateTransactionRequest{
		OrderID:       orderID,
		Amount:        int64(grandtotal),
		Items:         items,
		CustomerName:  req.Name,
		CustomerEmail: req.Email,
		CustomerPhone: req.Phone,
		Notes:         req.Notes,
	}

	midtransRes, err := t.midtransService.CreateTransaction(midtransReq)
	if err != nil {
		log.Errorf("[TransactionController] CreateTransaction - 3: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create transaction",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Transaction created successfully",
		"data": fiber.Map{
			"transaction_id": idTransaction,
			"payment_token":  midtransRes.PaymentToken,
			"order_id":       midtransRes.OrderID,
		},
	})

}

// GetDashboardByMerchant implements TransactionControllerInterface.
func (t *transactionController) GetDashboardByMerchant(c *fiber.Ctx) error {
	ctx := c.Context()

	merchantIDStr := c.Params("merchant_id")
	merchantID := conv.StringToUint(merchantIDStr)

	userIDStr := c.Query("user_id")
	userID := conv.StringToUint(userIDStr)

	totalRevenue, totalTransactions, productsSold, err := t.transactionUsecase.GetDashboardStatsByMerchant(ctx, userID, merchantID)
	if err != nil {
		log.Errorf("[TransactionController] GetDashboardByMerchant - 1: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get dashboard stats by merchant",
		})
	}

	response := response.DashboardByMerchantResponse{
		DashboardResponse: response.DashboardResponse{
			TotalRevenue:      totalRevenue,
			TotalTransactions: totalTransactions,
			ProductsSold:      productsSold,
		},
		Merchant: response.MerchantSummary{
			ID: merchantID,
		},
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    response,
		"message": "Dashboard stats fetched successfully",
	})
}

// GetManagerDashboard implements TransactionControllerInterface.
func (t *transactionController) GetManagerDashboard(c *fiber.Ctx) error {
	ctx := c.Context()
	userIDStr := c.Query("user_id")
	userID := conv.StringToUint(userIDStr)

	totalRevenue, totalTransactions, productsSold, err := t.transactionUsecase.GetDashboardStats(ctx, userID)
	if err != nil {
		log.Errorf("[TransactionController] GetManagerDashboard - 1: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get transactions",
		})
	}

	response := response.DashboardResponse{
		TotalRevenue:      totalRevenue,
		TotalTransactions: totalTransactions,
		ProductsSold:      productsSold,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    response,
		"message": "Dashboard stats fetched successfully",
	})
}

// GetTransactions implements TransactionControllerInterface.
func (t *transactionController) GetTransactions(c *fiber.Ctx) error {
	ctx := c.Context()

	query := request.GetAllTransactionRequest{}
	if err := c.QueryParser(&query); err != nil {
		log.Errorf("[TransactionController] GetTransactions - 1: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.Limit <= 0 {
		query.Limit = 10
	}

	merchantID := conv.StringToUint(query.MerchantID)

	transactions, total, err := t.transactionUsecase.GetTransactions(ctx, query.Page, query.Limit, query.Search, query.SortBy, query.SortOrder, merchantID)
	if err != nil {
		log.Errorf("[TransactionController] GetTransactions - 2: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get transactions",
		})
	}

	var transactionResponses []response.TransactionResponse
	for _, transaction := range transactions {
		var transactionProductResponses []response.TransactionProductResponse
		for _, tp := range transaction.TransactionProducts {
			transactionProductResponses = append(transactionProductResponses, response.TransactionProductResponse{
				ID:            tp.ID,
				ProductID:     tp.ProductID,
				ProductName:   tp.ProductName,
				ProductPhoto:  tp.ProductPhoto,
				ProductAbout:  tp.ProductAbout,
				Quantity:      tp.Quantity,
				Price:         tp.Price,
				SubTotal:      tp.SubTotal,
				TransactionID: tp.TransactionID,
				Category: struct {
					ID    uint   `json:"id"`
					Name  string `json:"name"`
					Photo string `json:"photo"`
				}{
					ID:    tp.ProductCategoryID,
					Name:  tp.ProductCategoryName,
					Photo: tp.ProductCategoryPhoto,
				},
			})
		}

		transactionResponses = append(transactionResponses, response.TransactionResponse{
			ID:                  transaction.ID,
			Name:                transaction.Name,
			Phone:               transaction.Phone,
			Email:               transaction.Email,
			Address:             transaction.Address,
			SubTotal:            transaction.SubTotal,
			TaxTotal:            transaction.TaxTotal,
			GrandTotal:          transaction.GrandTotal,
			MerchantID:          transaction.MerchantID,
			MerchantName:        transaction.MerchantName,
			PaymentStatus:       transaction.PaymentStatus,
			PaymentMethod:       transaction.PaymentMethod,
			TransactionCode:     transaction.TransactionCode,
			OrderID:             transaction.OrderID,
			Notes:               transaction.Notes,
			TransactionProducts: transactionProductResponses,
		})
	}

	paginationInfo := pagination.CalculatePagination(query.Page, query.Limit, int(total))

	response := response.GetAllTransactionsResponse{
		Transactions: transactionResponses,
		Pagination:   paginationInfo,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    response,
		"message": "Transactions fetched successfully",
	})
}

// GetTransactionByID implements TransactionControllerInterface.
func (t *transactionController) GetTransactionByID(c *fiber.Ctx) error {
	ctx := c.Context()

	idStr := c.Params("id")
	id := conv.StringToUint(idStr)

	if id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction ID",
		})
	}

	transaction, err := t.transactionUsecase.GetTransactionByID(ctx, id)
	if err != nil {
		log.Errorf("[TransactionController] GetTransactionByID - 1: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get transaction",
		})
	}

	var transactionProductResponses []response.TransactionProductResponse
	for _, tp := range transaction.TransactionProducts {
		transactionProductResponses = append(transactionProductResponses, response.TransactionProductResponse{
			ID:            tp.ID,
			ProductID:     tp.ProductID,
			ProductName:   tp.ProductName,
			ProductPhoto:  tp.ProductPhoto,
			ProductAbout:  tp.ProductAbout,
			Quantity:      tp.Quantity,
			Price:         tp.Price,
			SubTotal:      tp.SubTotal,
			TransactionID: tp.TransactionID,
			Category: struct {
				ID    uint   `json:"id"`
				Name  string `json:"name"`
				Photo string `json:"photo"`
			}{
				ID:    tp.ProductCategoryID,
				Name:  tp.ProductCategoryName,
				Photo: tp.ProductCategoryPhoto,
			},
		})
	}

	transactionResponse := response.TransactionResponse{
		ID:                  transaction.ID,
		Name:                transaction.Name,
		Phone:               transaction.Phone,
		Email:               transaction.Email,
		Address:             transaction.Address,
		SubTotal:            transaction.SubTotal,
		TaxTotal:            transaction.TaxTotal,
		GrandTotal:          transaction.GrandTotal,
		MerchantID:          transaction.MerchantID,
		MerchantName:        transaction.MerchantName,
		PaymentStatus:       transaction.PaymentStatus,
		PaymentMethod:       transaction.PaymentMethod,
		TransactionCode:     transaction.TransactionCode,
		OrderID:             transaction.OrderID,
		Notes:               transaction.Notes,
		TransactionProducts: transactionProductResponses,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    transactionResponse,
		"message": "Transaction fetched successfully",
	})
}

// MidtransCallback implements TransactionControllerInterface.
func (t *transactionController) MidtransCallback(c *fiber.Ctx) error {
	ctx := c.Context()

	req := request.MidtransCallbackRequest{}
	if err := c.BodyParser(&req); err != nil {
		log.Errorf("[TransactionController] MidtransCallback - 1: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Konversi status Midtrans ke konstanta internal
	internalStatus := model.ConvertMidtransStatusToInternal(req.TransactionStatus)
	log.Infof("[TransactionController] MidtransCallback - Converting status: %s -> %s for order_id: %s", req.TransactionStatus, internalStatus, req.OrderID)

	if err := t.transactionUsecase.UpdatePaymentStatus(ctx, req.OrderID, internalStatus, req.PaymentType, req.TransactionID, req.FraudStatus); err != nil {
		log.Errorf("[TransactionController] MidtransCallback - 2: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update payment status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment status updated successfully",
	})
}

func NewTransactionController(transactionUsecase usecase.TransactionUsecaseInterface, midtransService midtrans.MidtransServiceInterface) TransactionControllerInterface {
	return &transactionController{
		transactionUsecase: transactionUsecase,
		midtransService:    midtransService,
	}
}
