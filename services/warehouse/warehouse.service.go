package warehouse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/services/warehouse/types/models"

	_ "github.com/lib/pq"
)

func GetWarehouseLocations(w http.ResponseWriter, r *http.Request) {
	DB := database.GetDB()
	ctx := r.Context()

	if err := DB.PingContext(ctx); err != nil {
		panic(err)
	}

	var warehouseLocationRows []models.WarehouseLocationModel

	dbResponse, err := DB.QueryContext(ctx,
		`SELECT id, name, location_id, skus_count from public.warehouse_locations ORDER BY id`)
	if err != nil {
		fmt.Println("Error in Querying warehouse_locations")
	}
	defer dbResponse.Close()

	for dbResponse.Next() {
		var wl models.WarehouseLocationModel
		var rawSkusCount json.RawMessage
		var sc models.Sku

		if err := dbResponse.Scan(&wl.ID, &wl.Name, &wl.LocationId, &rawSkusCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(rawSkusCount, &sc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		wl.SkusCount = sc

		warehouseLocationRows = append(warehouseLocationRows, wl)
	}

	helper.WriteJson(w, http.StatusOK, warehouseLocationRows)

}
