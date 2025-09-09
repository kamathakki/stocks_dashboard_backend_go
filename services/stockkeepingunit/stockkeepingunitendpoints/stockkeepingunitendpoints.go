package stockkeepingunitendpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/database/redis"
	"stock_automation_backend_go/helper"
	"stockkeepingunit/types/models"
	"stockkeepingunit/types/models/responsemodels"
	"strings"
	"time"
	updatestockcountforwarehouselocationpb "warehouse/proto"

	"github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)



func GetStockKeepingUnits(w http.ResponseWriter, r *http.Request) ([]responsemodels.StockKeepingUnit, error) {
	DB := database.GetDB()
	ctx := r.Context()

	dbResponse, err := DB.QueryContext(ctx, `SELECT id, name, sku_code, excel_names, model_no, weight, carton_weight, fitting_in_carton
                    FROM stockkeepingunit.stock_keeping_units WHERE is_deleted = false ORDER BY id`)

	if err != nil {
		return nil, err
	}

	var SKURow responsemodels.StockKeepingUnit
	var stockKeepingUnitRows []responsemodels.StockKeepingUnit

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
  var stockCountBody struct { StockCount models.StockCountByWarehouseCountries `json:"stockCount"` }
  if err := json.NewDecoder(r.Body).Decode(&stockCountBody.StockCount); err != nil {
	return false, fmt.Errorf("error decoding stock count body: %w", err)
  }
  defer r.Body.Close()

  parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
  if(len(parts) < 2) {
	return false, fmt.Errorf("country id is required")
  }
  countryIdStr := parts[len(parts)-1]

  stockCountStr, err := json.Marshal(stockCountBody.StockCount)
  if err != nil {
	return false, fmt.Errorf("error marshalling stock count: %w", err)
  }

  if err := redis.SetHash("stockcount", countryIdStr, string(stockCountStr)); err != nil {
	return false, err
  }

  return true, nil
}

func UpdateStockCountByCountry(w http.ResponseWriter, r *http.Request) (any, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:7002",
     grpc.WithTransportCredentials(insecure.NewCredentials()))
	 if err != nil {
		fmt.Println("Error connecting to warehouse gRPC server:", err)
		return false, err
	 }
	 defer conn.Close()
	 
	grpcClient := updatestockcountforwarehouselocationpb.NewWarehouseClient(conn)

	var stockCountBody struct { StockCount models.StockCountByWarehouseCountries `json:"stockCount"` }
	if err := json.NewDecoder(r.Body).Decode(&stockCountBody); err != nil {
		return false, err
	}

	defer r.Body.Close()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if(len(parts) < 2) {
	  return false, fmt.Errorf("country id is required")
	}
	countryIdStr := parts[len(parts)-1]

	fmt.Println("Reached here mother")

	warehouseLocationStructure, err := redis.GetKey[[]models.WarehouseStructure]("warehouseLocationsStructure")
	if err != nil {
	  return false, err
	}
	fmt.Println("Reached here mother 1")
	
	for _, warehouse := range stockCountBody.StockCount.Warehouses {
		fmt.Println("Reached here mother 2")
		warehouseStructure, _ := helper.FindByWhere(warehouseLocationStructure, func(w models.WarehouseStructure) string { return w.Name }, warehouse.Name)
		if warehouseStructure == nil {
			return false, fmt.Errorf("warehouse %s not found in cached structure", warehouse.Name)
		}

		for _, location := range warehouse.Locations {
			fmt.Println("Reached here mother 3")
			_, locationIdx := helper.FindByWhere(&warehouseStructure.Locations, func(l models.WarehouseLocationEntry) int { return l.LocationId }, location.LocationId)
			if locationIdx < 0 || locationIdx >= len(warehouseStructure.WarehouseLocationIds) {
				return false, fmt.Errorf("location %d not found for warehouse %s", location.LocationId, warehouse.Name)
			}

			warehouseLocationId := warehouseStructure.WarehouseLocationIds[locationIdx]

			skuForLocation, ok := warehouse.Sku[location.LocationName]
			if !ok {
				return false, fmt.Errorf("sku not provided for location %s in warehouse %s", location.LocationName, warehouse.Name)
			}


			fmt.Println("Reached here mother 4", warehouseStructure.ID, warehouseLocationId)

			res, err := grpcClient.UpdateStockcountForWarehouselocation(ctx, &updatestockcountforwarehouselocationpb.StockCountUpdateRequest{
				WarehouseId: int64(warehouseStructure.ID),
				WarehouseLocationId: int64(warehouseLocationId),
				StockCount: map[string]int64{
					"GKS":     int64(skuForLocation.GKS),
					"NEO":     int64(skuForLocation.NEO),
					"PRO":     int64(skuForLocation.PRO),
					"MM-B":    int64(skuForLocation.MMB),
					"MM-P":    int64(skuForLocation.MMP),
					"SWAP":    int64(skuForLocation.SWAP),
					"M3-MR":   int64(skuForLocation.M3MR),
					"M3-PB":   int64(skuForLocation.M3PB),
					"M3-FIFA": int64(skuForLocation.M3FIFA),
				},
			})
			if err != nil {
				return false, err
			}
			if !res.Updated {
				return false, fmt.Errorf("failed to update stock count for warehouse location %d", warehouseLocationId)
			}
			fmt.Println("Reached here mother 5")
		}
	}

	if err := redis.DeleteHash("stockcount", countryIdStr); err != nil {
		fmt.Printf("Error %v deleting stockcount for countryId: %s", err, countryIdStr)
	}


	return true, nil
}