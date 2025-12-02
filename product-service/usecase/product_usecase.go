package usecase

import (
	"context"
	"errors"
	"micro-warehouse/product-service/model"
	"micro-warehouse/product-service/pkg/httpclient"
	"micro-warehouse/product-service/repository"

	"github.com/gofiber/fiber/v2/log"
)

type ProductUsecaseInterface interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	GetAllProducts(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]model.Product, int64, error)
	GetProductByID(ctx context.Context, id uint) (*model.Product, error)
	GetProductByBarcode(ctx context.Context, barcode string) (*model.Product, error)
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, id uint) error
}

type productUsecase struct {
	productRepo     repository.ProductRepositoryInterface
	warehouseClient *httpclient.WarehouseClient
	merchantClient  *httpclient.MerchantClient
}

// CreateProduct implements ProductUsecaseInterface.
func (p *productUsecase) CreateProduct(ctx context.Context, product *model.Product) error {
	return p.productRepo.CreateProduct(ctx, product)
}

// DeleteProduct implements ProductUsecaseInterface.
func (p *productUsecase) DeleteProduct(ctx context.Context, id uint) error {
	warehouseStock, err := p.warehouseClient.GetProductStockAcrossWarehouses(ctx, id)
	if err != nil {
		log.Errorf("[DeleteProduct] Failed to check warehouse stock for product %d", id)
		return err
	}

	if warehouseStock > 0 {
		log.Errorf("[DeleteProduct] Product %d has stock in warehouse", id)
		return errors.New("product has stock in warehouse")
	}

	merchantStock, err := p.merchantClient.GetProductStockAcrossMerchants(ctx, id)
	if err != nil {
		log.Errorf("[DeleteProduct] Failed to check merchant stock for product %d", id)
		return err
	}

	if merchantStock > 0 {
		log.Errorf("[DeleteProduct] Product %d has stock in merchant", id)
		return errors.New("product has stock in merchant")
	}

	if err := p.merchantClient.DeleteAllProductMerchantProducts(ctx, id); err != nil {
		log.Errorf("[DeleteProduct] Failed to delete all merchant product for product %d", id)
		return err
	}

	if err := p.warehouseClient.DeleteAllProductWarehouseProducts(ctx, id); err != nil {
		log.Errorf("[DeleteProduct] Failed to delete all warehouse product for product %d", id)
		return err
	}

	return p.productRepo.DeleteProduct(ctx, id)
}

// GetAllProducts implements ProductUsecaseInterface.
func (p *productUsecase) GetAllProducts(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]model.Product, int64, error) {
	return p.productRepo.GetAllProducts(ctx, page, limit, search, sortBy, sortOrder)
}

// GetProductByBarcode implements ProductUsecaseInterface.
func (p *productUsecase) GetProductByBarcode(ctx context.Context, barcode string) (*model.Product, error) {
	return p.productRepo.GetProductByBarcode(ctx, barcode)
}

// GetProductByID implements ProductUsecaseInterface.
func (p *productUsecase) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	return p.productRepo.GetProductByID(ctx, id)
}

// UpdateProduct implements ProductUsecaseInterface.
func (p *productUsecase) UpdateProduct(ctx context.Context, product *model.Product) error {
	return p.productRepo.UpdateProduct(ctx, product)
}

func NewProductUsecase(productRepo repository.ProductRepositoryInterface) ProductUsecaseInterface {
	return &productUsecase{productRepo: productRepo}
}
