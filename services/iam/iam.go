package iam

import (
	"fmt"
	"net/http"
	"stock_automation_backend_go/helper"

	// routes "stock_automation_backend_go/services/api-gateway/api-routes"
	"stock_automation_backend_go/shared/env"
)

var localmux *http.ServeMux

type ResponseStruct struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("IAM service up"))
}

func health(w http.ResponseWriter, r *http.Request) {
	res := ResponseStruct{StatusCode: http.StatusOK, Message: "OK"}
	helper.WriteJson(w, http.StatusOK, res, nil)
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

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)

	// mux.HandleFunc("/api/warehouse/getWarehouseLocations", responseWrapper(warehouse.GetWarehouseLocations))
	// mux.HandleFunc("/api/stockkeepingunit/getStockKeepingUnits", responseWrapper(stockkeepingunit.GetStockKeepingUnits))
	mux.HandleFunc("/api/iam/login", responseWrapper(Login))
	localmux = mux
}

func main() {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.IAM_PORT)),
		Handler: localmux,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
