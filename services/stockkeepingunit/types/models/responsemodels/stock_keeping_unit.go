package responsemodels

type StockKeepingUnit struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	SkuCode         string   `json:"skuCode"`
	ExcelNames      []string `json:"excelNames"`
	ModelNo         string   `json:"modelNo"`
	Weight          float64  `json:"weight"`
	CartonWeight    float64  `json:"cartonWeight"`
	FittingInCarton int64    `json:"fittingInCarton"`
}
