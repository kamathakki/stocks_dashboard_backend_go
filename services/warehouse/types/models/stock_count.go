package models

type StockCountByWarehouseCountries struct {
	CountryId   int         `json:"countryId"`
	CountryName string      `json:"countryName"`
	Warehouses  []Warehouse `json:"warehouses"`
}

type Warehouse struct {
	ID        int                      `json:"id"`
	Name      string                   `json:"name"`
	Locations []WarehouseLocationEntry `json:"locations"`
	Sku       map[string]Sku           `json:"sku"`
}

type WarehouseLocationEntry struct {
	LocationName string `json:"locationName"`
	LocationId   int    `json:"locationId"`
}

type Sku struct {
	GKS    int `json:"GKS"`
	NEO    int `json:"NEO"`
	PRO    int `json:"PRO"`
	MMB    int `json:"MM-B"`
	MMP    int `json:"MM-P"`
	SWAP   int `json:"SWAP"`
	M3MR   int `json:"M3-MR"`
	M3PB   int `json:"M3-PB"`
	M3FIFA int `json:"M3-FIFA"`
}

type WarehouseStructure struct {
	ID                   int                      `json:"id"`
	Name                 string                   `json:"name"`
	Countries            []string                 `json:"countries"`
	Locations            []WarehouseLocationEntry `json:"locations"`
	CountryIds           []int                    `json:"countryIds"`
	WarehouseLocationIds []int                    `json:"warehouseLocationIds"`
	CountryLocations     map[string][]string      `json:"countryLocations"`
}
