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
	mux.HandleFunc("/getColumnMappings/", common.APIWrapper(warehouseendpoints.GetColumnMappings))
	mux.HandleFunc("/getWarehouseLocationsStructure", common.APIWrapper(warehouseendpoints.GetWarehouseLocationsStructure))
	mux.HandleFunc("/getCountries", common.APIWrapper(warehouseendpoints.GetCountries))
	mux.HandleFunc("/getStockCountFromHistory/", common.APIWrapper(warehouseendpoints.GetStockCountFromHistory))
	mux.HandleFunc("/addStockCountHistoryForCountry/", common.APIWrapper(warehouseendpoints.AddStockCountHistoryForCountry))
	mux.HandleFunc("/getStockCountByWarehouseCountries", common.APIWrapper(warehouseendpoints.GetStockCountByWarehouseCountries))
	mux.HandleFunc("/updateWarehouseColumnMapping", common.APIWrapper(warehouseendpoints.UpdateWarehouseColumnMapping))
	mux.HandleFunc("/deleteWarehouseColumnMapping/", common.APIWrapper(warehouseendpoints.DeleteWarehouseColumnMapping))
	mux.HandleFunc("/getStockCountData/", common.APIWrapper(warehouseendpoints.GetStockCountData))
	mux.HandleFunc("/updateStockCountForWarehouseLocation/", common.APIWrapper(warehouseendpoints.UpdateStockCountForWarehouseLocation))
	return restrictedmux
}
