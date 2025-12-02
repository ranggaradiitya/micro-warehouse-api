package repository

import (
	"context"
	"micro-warehouse/transaction-service/model"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

// get data overview dashboard manager/keeper,create transaction, update status transaction
type TransactionRepositoryInterface interface {
	GetDashboardStats(ctx context.Context) (int64, int64, int64, error)
	GetDashboardStatsByMerchant(ctx context.Context, merchantID uint) (int64, int64, int64, error)

	GetTransactions(ctx context.Context, page, limit int, search, sortBy, sortOrder string, merchantID uint) ([]model.Transaction, int64, error)
	GetTransactionByID(ctx context.Context, id uint) (*model.Transaction, error)
	CreateTransaction(ctx context.Context, transaction model.Transaction) (int64, error)

	// Midtrans update status transaction
	UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus, paymentMethod, transactionID, fraudStatus string) error
}

type transactionRepository struct {
	db *gorm.DB
}

// CreateTransaction implements TransactionRepositoryInterface.
func (t *transactionRepository) CreateTransaction(ctx context.Context, transaction model.Transaction) (int64, error) {
	select {
	case <-ctx.Done():
		log.Errorf("[TransactionRepository] CreateTransaction - 1: %v", ctx.Err())
		return 0, ctx.Err()
	default:
		tx := t.db.WithContext(ctx).Begin()
		if tx.Error != nil {
			log.Errorf("[TransactionRepository] CreateTransaction - 2: %v", tx.Error)
			return 0, tx.Error
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Errorf("[TransactionRepository] CreateTransaction - 3: %v", r)
			}
		}()

		products := transaction.TransactionProducts
		transaction.TransactionProducts = nil

		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			log.Errorf("[TransactionRepository] CreateTransaction - 4: %v", err)
			return 0, err
		}

		for _, product := range products {
			modelTransactionProduct := model.TransactionProduct{
				ProductID:     product.ProductID,
				Quantity:      product.Quantity,
				Price:         product.Price,
				SubTotal:      product.SubTotal,
				TransactionID: transaction.ID,
			}

			if err := tx.Create(&modelTransactionProduct).Error; err != nil {
				tx.Rollback()
				log.Errorf("[TransactionRepository] CreateTransaction - 5: %v", err)
				return 0, err
			}
		}

		if err := tx.Commit().Error; err != nil {
			log.Errorf("[TransactionRepository] CreateTransaction - 6: %v", err)
			return 0, err
		}

		return int64(transaction.ID), nil
	}
}

// GetDashboardStats implements TransactionRepositoryInterface.
func (t *transactionRepository) GetDashboardStats(ctx context.Context) (int64, int64, int64, error) {
	select {
	case <-ctx.Done():
		log.Errorf("[TransactionRepository] GetDashboardStats - 1: %v", ctx.Err())
		return 0, 0, 0, ctx.Err()
	default:
		var totalRevenue int64
		var totalTransactions int64
		var productsSold int64

		var result struct {
			TotalRevenue      int64 `json:"total_revenue"`
			TotalTransactions int64 `json:"total_transactions"`
		}

		err := t.db.WithContext(ctx).Model(&model.Transaction{}).
			Where("payment_status = ?", model.PaymentStatusSuccess).
			Select("COALESCE(SUM(grand_total), 0) as total_revenue, COUNT(*) as total_transactions").
			Scan(&result).Error

		if err != nil {
			log.Errorf("[TransactionRepository] GetDashboardStats - 2: %v", err)
			return 0, 0, 0, err
		}

		totalRevenue = result.TotalRevenue
		totalTransactions = result.TotalTransactions

		err = t.db.WithContext(ctx).Model(&model.TransactionProduct{}).
			Joins("JOIN transactions ON transaction_products.transaction_id = transactions.id").
			Where("transactions.payment_status = ?", model.PaymentStatusSuccess).
			Select("COALESCE(SUM(transaction_products.quantity), 0) as products_sold").
			Scan(&productsSold).Error

		if err != nil {
			log.Errorf("[TransactionRepository] GetDashboardStats - 3: %v", err)
			return 0, 0, 0, err
		}

		return totalRevenue, totalTransactions, productsSold, nil

	}
}

// GetDashboardStatsByMerchant implements TransactionRepositoryInterface.
func (t *transactionRepository) GetDashboardStatsByMerchant(ctx context.Context, merchantID uint) (int64, int64, int64, error) {
	select {
	case <-ctx.Done():
		log.Errorf("[TransactionRepository] GetDashboardStatsByMerchant - 1: %v", ctx.Err())
		return 0, 0, 0, ctx.Err()
	default:
		if merchantID == 0 {
			return 0, 0, 0, nil
		}

		var totalRevenue int64
		var totalTransactions int64
		var productsSold int64

		var result struct {
			TotalRevenue      int64 `json:"total_revenue"`
			TotalTransactions int64 `json:"total_transactions"`
		}

		err := t.db.WithContext(ctx).Model(&model.Transaction{}).
			Where("merchant_id = ? AND payment_status = ?", merchantID, model.PaymentStatusSuccess).
			Select("COALESCE(SUM(grand_total), 0) as total_revenue, COUNT(*) as total_transactions").
			Scan(&result).Error

		if err != nil {
			log.Errorf("[TransactionRepository] GetDashboardStatsByMerchant - 2: %v", err)
			return 0, 0, 0, err
		}

		totalRevenue = result.TotalRevenue
		totalTransactions = result.TotalTransactions

		err = t.db.WithContext(ctx).Model(&model.TransactionProduct{}).
			Joins("JOIN transactions ON transaction_products.transaction_id = transactions.id").
			Where("transactions.merchant_id = ? AND transactions.payment_status = ?", merchantID, model.PaymentStatusSuccess).
			Select("COALESCE(SUM(transaction_products.quantity), 0) as products_sold").
			Scan(&productsSold).Error

		if err != nil {
			log.Errorf("[TransactionRepository] GetDashboardStatsByMerchant - 3: %v", err)
			return 0, 0, 0, err
		}

		return totalRevenue, totalTransactions, productsSold, nil
	}
}

// GetTransactions implements TransactionRepositoryInterface.
func (t *transactionRepository) GetTransactions(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, merchantID uint) ([]model.Transaction, int64, error) {
	select {
	case <-ctx.Done():
		log.Errorf("[TransactionRepository] GetTransactions - 1: %v", ctx.Err())
		return nil, 0, ctx.Err()
	default:
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 10
		}
		if sortBy == "" {
			sortBy = "created_at"
		}
		if sortOrder == "" {
			sortOrder = "desc"
		}

		offset := (page - 1) * limit

		baseSql := t.db.WithContext(ctx).Model(&model.Transaction{}).
			Preload("TransactionProducts")

		if search != "" {
			searchTerm := "%" + search + "%"
			baseSql = baseSql.Where("name ILIKE ? OR phone ILIKE ?",
				searchTerm, searchTerm)
		}

		if merchantID != 0 {
			baseSql = baseSql.Where("merchant_id = ?", merchantID)
		}

		var totalRecords int64
		if err := baseSql.Count(&totalRecords).Error; err != nil {
			log.Errorf("[GetAllTransactions - Count error] Failed to count transactions")
			return nil, 0, err
		}

		var transactions []model.Transaction
		err := baseSql.WithContext(ctx).
			Preload("TransactionProducts").
			Order(sortBy + " " + sortOrder).
			Offset(offset).
			Limit(limit).
			Find(&transactions).Error

		if err != nil {
			log.Errorf("[TransactionRepository] GetTransactions - 2: %v", err)
			return nil, 0, err
		}
		return transactions, totalRecords, nil
	}
}

// GetTransactionByID implements TransactionRepositoryInterface.
func (t *transactionRepository) GetTransactionByID(ctx context.Context, id uint) (*model.Transaction, error) {
	select {
	case <-ctx.Done():
		log.Errorf("[TransactionRepository] GetTransactionByID - 1: %v", ctx.Err())
		return nil, ctx.Err()
	default:
		var transaction model.Transaction
		err := t.db.WithContext(ctx).
			Preload("TransactionProducts").
			Where("id = ?", id).
			First(&transaction).Error

		if err != nil {
			log.Errorf("[TransactionRepository] GetTransactionByID - 2: %v", err)
			return nil, err
		}

		return &transaction, nil
	}
}

// UpdatePaymentStatus implements TransactionRepositoryInterface.
func (t *transactionRepository) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string, paymentMethod string, transactionID string, fraudStatus string) error {
	select {
	case <-ctx.Done():
		log.Errorf("[TransactionRepository] UpdatePaymentStatus - 1: %v", ctx.Err())
		return ctx.Err()
	default:

		if err := t.db.WithContext(ctx).Model(&model.Transaction{}).Where("order_id = ?", orderID).First(&model.Transaction{}).Error; err != nil {
			log.Errorf("[TransactionRepository] UpdatePaymentStatus - 2: %v", err)
			return err
		}

		updates := map[string]interface{}{
			"payment_status": paymentStatus,
		}

		if paymentMethod != "" {
			updates["payment_method"] = paymentMethod
		}
		if transactionID != "" {
			updates["transaction_code"] = transactionID
		}
		if fraudStatus != "" {
			updates["fraud_status"] = fraudStatus
		}

		err := t.db.WithContext(ctx).
			Model(&model.Transaction{}).
			Where("order_id = ?", orderID).
			Updates(updates).Error

		if err != nil {
			log.Errorf("[TransactionRepository] UpdatePaymentStatus - 2: %v", err)
			return err
		}

		return nil
	}
}

func NewTransactionRepository(db *gorm.DB) TransactionRepositoryInterface {
	return &transactionRepository{db: db}
}
