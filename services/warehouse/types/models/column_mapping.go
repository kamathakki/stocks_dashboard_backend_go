package models

type ColumnMapping struct {
	StandardColumnId         int    `json:"standardColumnId"`
	WarehouseColumnMappingId int    `json:"warehouseColumnMappingId"`
	StandardName             string `json:"standardName"`
	ExcelName                string `json:"excelName"`
}
