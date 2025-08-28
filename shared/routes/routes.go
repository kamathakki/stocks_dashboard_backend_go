package routes

import (
	"net/http"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/services/iam"
	"stock_automation_backend_go/services/stockkeepingunit"
	"stock_automation_backend_go/services/warehouse"
)

func responseWrapper[T any](handler func(w http.ResponseWriter, r *http.Request) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := handler(w, r)
		if err != nil {
			helper.WriteJson(w, http.StatusInternalServerError, nil, err)
		} else {
			helper.WriteJson(w, http.StatusOK, result, nil)
		}
	}
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/warehouse/getWarehouseLocations", responseWrapper(warehouse.GetWarehouseLocations))
	mux.HandleFunc("/api/stockkeepingunit/getStockKeepingUnits", responseWrapper(stockkeepingunit.GetStockKeepingUnits))
	mux.HandleFunc("/api/iam/login", responseWrapper(iam.Login))
}
