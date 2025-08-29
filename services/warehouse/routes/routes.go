package routes

import (
	"iam/env"
	"net/http"
	"stock_automation_backend_go/helper"
	common "stock_automation_backend_go/shared"
	"warehouse/warehouseendpoints"
)

type ResponseStruct struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func requireInternal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Port() != env.GetEnv("BACKEND_PORT") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Warehouse service up"))
}

func health(w http.ResponseWriter, r *http.Request) {
	res := ResponseStruct{StatusCode: http.StatusOK, Message: "OK"}
	helper.WriteJson(w, http.StatusOK, res, nil)
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)

	// mux.HandleFunc("/api/warehouse/getWarehouseLocations", responseWrapper(warehouse.GetWarehouseLocations))
	// mux.HandleFunc("/api/stockkeepingunit/getStockKeepingUnits", responseWrapper(stockkeepingunit.GetStockKeepingUnits))
	mux.HandleFunc("/getWarehouseLocations", common.ResponseWrapper(warehouseendpoints.GetWarehouseLocations))
}
