package routes

import (
	"net/http"
	"stock_automation_backend_go/services/iam"
	"stock_automation_backend_go/services/stockkeepingunit"
	"stock_automation_backend_go/services/warehouse"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/warehouse/getWarehouseLocations", warehouse.GetWarehouseLocations)
	mux.HandleFunc("/api/stockkeepingunit/getStockKeepingUnits", stockkeepingunit.GetStockKeepingUnits)
	mux.HandleFunc("/api/iam/login", iam.Login)
}
