package routes

import (
	"net/http"
	"stock_automation_backend_go/helper"
	common "stock_automation_backend_go/shared"
	"stockkeepingunit/stockkeepingunitendpoints"
)

type ResponseStruct struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SKU service up"))
}

func health(w http.ResponseWriter, r *http.Request) {
	res := ResponseStruct{StatusCode: http.StatusOK, Message: "OK"}
	helper.WriteJson(w, http.StatusOK, res, nil)
}

func RegisterRoutes(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)

	// mux.HandleFunc("/api/warehouse/getWarehouseLocations", responseWrapper(warehouse.GetWarehouseLocations))
	// mux.HandleFunc("/api/stockkeepingunit/getStockKeepingUnits", responseWrapper(stockkeepingunit.GetStockKeepingUnits))
	restrictedmux := common.RequireInternal(mux)
	mux.HandleFunc("/getStockKeepingUnits", common.APIWrapper(stockkeepingunitendpoints.GetStockKeepingUnits))
	mux.HandleFunc("/updateStockCountInMemory/", common.APIWrapper(stockkeepingunitendpoints.UpdateStockCountInMemory))
	mux.HandleFunc("/updateStockCountByCountry/", common.APIWrapper(stockkeepingunitendpoints.UpdateStockCountByCountry))
	return restrictedmux
}
