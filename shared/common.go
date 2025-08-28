package common

import (
	"net/http"
	"stock_automation_backend_go/helper"
)

func ResponseWrapper[T any](handler func(w http.ResponseWriter, r *http.Request) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := handler(w, r)
		if err != nil {
			helper.WriteJson(w, http.StatusInternalServerError, nil, err)
		} else {
			helper.WriteJson(w, http.StatusOK, result, nil)
		}
	}
}

func APIWrapper[T any](handler func(w http.ResponseWriter, r *http.Request) (T, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := handler(w, r)
		if err != nil {
			helper.WriteMicroServiceJson(w, http.StatusInternalServerError, nil, err)
		} else {
			helper.WriteMicroServiceJson(w, http.StatusOK, result, nil)
		}

	}
}

func RequireInternal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromgateway := r.Header.Get("fromgateway")

		if fromgateway != "y" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
