package models

type StockKeepingUnit struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	SkuCode         string   `json:"sku_code"`
	ExcelNames      []string `json:"excel_names"`
	ModelNo         string   `json:"model_no"`
	Weight          float64  `json:"weight"`
	CartonWeight    float64  `json:"carton_weight"`
	FittingInCarton int64    `json:"fitting_in_carton"`
}
