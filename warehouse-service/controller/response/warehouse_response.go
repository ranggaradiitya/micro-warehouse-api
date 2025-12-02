package response

import "micro-warehouse/warehouse-service/pkg/pagination"

type WarehouseResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	Photo        string `json:"photo"`
	Phone        string `json:"phone"`
	CountProduct int    `json:"count_product"`
}

type GetAllWarehouseResponse struct {
	Warehouses []WarehouseResponse           `json:"warehouses"`
	Pagination pagination.PaginationResponse `json:"pagination"`
}

type DetailWarehouseResponse struct {
	ID                uint                       `json:"id"`
	Name              string                     `json:"name"`
	Address           string                     `json:"address"`
	Photo             string                     `json:"photo"`
	Phone             string                     `json:"phone"`
	WarehouseProducts []WarehouseProductResponse `json:"warehouse_products"`
}
