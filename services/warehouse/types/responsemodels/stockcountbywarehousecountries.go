package responsemodels

import "warehouse/types/models"

type StockCountByWarehouseCountries struct {
	CountryId   int         `json:"countryId"`
	CountryName string      `json:"countryName"`
	Warehouses  []Warehouse `json:"warehouses"`
}

type Warehouse struct {
	ID        int                             `json:"id"`
	Name      string                          `json:"name"`
	Locations []models.WarehouseLocationEntry `json:"locations"`
	Sku       map[string]models.Sku           `json:"sku"`
}
