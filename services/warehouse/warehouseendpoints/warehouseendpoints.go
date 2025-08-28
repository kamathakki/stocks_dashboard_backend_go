package warehouseendpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/database"
	"warehouse/types/models"

	_ "github.com/lib/pq"
)

func GetWarehouseLocations(w http.ResponseWriter, r *http.Request) ([]models.WarehouseLocationModel, error) {
	DB := database.GetDB()
	ctx := r.Context()

	var warehouseLocationRows []models.WarehouseLocationModel

	dbResponse, err := DB.QueryContext(ctx,
		`SELECT id, name, location_id, skus_count from warehouse.warehouse_locations ORDER BY id`)
	if err != nil {
		fmt.Println("Error in Querying warehouse_locations")
	}
	defer dbResponse.Close()

	for dbResponse.Next() {
		var wl models.WarehouseLocationModel
		var rawSkusCount json.RawMessage
		var sc models.Sku

		if err := dbResponse.Scan(&wl.ID, &wl.Name, &wl.LocationId, &rawSkusCount); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rawSkusCount, &sc); err != nil {
			return nil, err
		}

		wl.SkusCount = sc

		warehouseLocationRows = append(warehouseLocationRows, wl)
	}

	return warehouseLocationRows, nil
}
