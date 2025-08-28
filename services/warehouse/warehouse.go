package warehouse

import (
	"fmt"
	"net/http"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/shared/env"
)

var localmux *http.ServeMux

type ResponseStruct struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

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

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("IAM service up"))
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
	mux.HandleFunc("/api/warehouse/getWarehouseLocations", responseWrapper(GetWarehouseLocations))
	localmux = mux
}

func main() {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.WAREHOUSE_PORT)),
		Handler: localmux,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
