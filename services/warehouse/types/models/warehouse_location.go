package models

type WarehouseLocationModel struct {
	ID         int    `json:"id"`
	LocationId int    `json:"location_id"`
	Name       string `json:"name"`
	SkusCount  Sku    `json:"skus_count"`
}
