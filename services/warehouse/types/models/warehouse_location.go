package models

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

type WarehouseLocationModel struct {
	ID         int    `json:"id"`
	LocationId int    `json:"location_id"`
	Name       string `json:"name"`
	SkusCount  Sku    `json:"skus_count"`
}
