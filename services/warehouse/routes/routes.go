package routes

import (
	"net/http"
	"stock_automation_backend_go/helper"
	common "stock_automation_backend_go/shared"
	"warehouse/warehouseendpoints"
)

type ResponseStruct struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Warehouse service up"))
}

func health(w http.ResponseWriter, r *http.Request) {
	res := ResponseStruct{StatusCode: http.StatusOK, Message: "OK"}
	helper.WriteJson(w, http.StatusOK, res, nil)
}

func RegisterRoutes(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)

	restrictedmux := common.RequireInternal(mux)

	mux.HandleFunc("/getWarehouseLocations", common.APIWrapper(warehouseendpoints.GetWarehouseLocations))
	// mux.HandleFunc("/api/stockkeepingunit/getStockKeepingUnits", responseWrapper(stockkeepingunit.GetStockKeepingUnits))
	return restrictedmux
}
