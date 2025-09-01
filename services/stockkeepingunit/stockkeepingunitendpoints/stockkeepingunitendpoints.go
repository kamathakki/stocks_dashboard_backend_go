package stockkeepingunitendpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/database/redis"
	"stockkeepingunit/types/models"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func GetStockKeepingUnits(w http.ResponseWriter, r *http.Request) ([]models.StockKeepingUnit, error) {
	DB := database.GetDB()
	ctx := r.Context()

	dbResponse, err := DB.QueryContext(ctx, `SELECT id, name, sku_code, excel_names, model_no, weight, carton_weight, fitting_in_carton
                    FROM stockkeepingunit.stock_keeping_units WHERE is_deleted = false ORDER BY id`)

	if err != nil {
		return nil, err
	}

	var SKURow models.StockKeepingUnit
	var stockKeepingUnitRows []models.StockKeepingUnit

	defer dbResponse.Close()

	for dbResponse.Next() {
		if err := dbResponse.Scan(&SKURow.ID, &SKURow.Name, &SKURow.SkuCode, pq.Array(&SKURow.ExcelNames), &SKURow.ModelNo,
			&SKURow.Weight, &SKURow.CartonWeight, &SKURow.FittingInCarton); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		stockKeepingUnitRows = append(stockKeepingUnitRows, SKURow)
	}

	//helper.WriteJson(w, http.StatusOK, stockKeepingUnitRows)
	return stockKeepingUnitRows, nil
}

func UpdateStockCountInMemory(w http.ResponseWriter, r *http.Request) (bool, error) {
  var stockCount models.StockCountByWarehouseCountries
  if err := json.NewDecoder(r.Body).Decode(&stockCount); err != nil {
	return false, err
  }
  defer r.Body.Close()

  parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
  if(len(parts) < 2) {
	return false, fmt.Errorf("country id is required")
  }
  countryIdStr := parts[len(parts)-1]

  stockCountStr, err := json.Marshal(stockCount)
  if err != nil {
	return false, err
  }

  if err := redis.SetHash("stockcount", countryIdStr, string(stockCountStr)); err != nil {
	return false, err
  }

  return true, nil
}

// func UpdateStockCountByCountry(w http.ResponseWriter, r *http.Request) (bool, error) {
//   var stockCount models.StockCountByWarehouseCountries
//   if err := json.NewDecoder(r.Body).Decode(&stockCount); err != nil {
// 	return false, err
//   }
//   defer r.Body.Close()
//   parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
//   if(len(parts) < 2) {
// 	return false, fmt.Errorf("country id is required")
//   }
//   countryIdStr := parts[len(parts)-1]

//   warehouseLocationStructure, err := redis.GetHash("warehouseLocationsStructure", countryIdStr, &[]models.WarehouseStructure{})
//   if err != nil {
// 	return false, err
//   }

//   for _, warehouse := range stockCount.Warehouses {
//       warehouseStructure, _ := helper.FindByWhere(warehouseLocationStructure.([]models.WarehouseStructure), func(w models.WarehouseStructure) string { return w.Name }, warehouse.Name) 

// 	  for _, location := range warehouse.Locations {
// 		_, locationIndex := helper.FindByWhere(warehouseStructure.Locations, func(l models.WarehouseLocationEntry) int { return l.LocationId }, location.LocationId)
// 		warehouseLocationId := warehouseStructure.WarehouseLocationIds[locationIndex]
// 		warehouseStructure.Locations = append(warehouseStructure.Locations, models.WarehouseLocationEntry{LocationName: location.LocationName, LocationId: location.LocationId})
// 	  }
//   }
// }