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
	"strings"
	"time"
	updatestockcountforwarehouselocationpb "warehouse/proto"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
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

func UpdateStockCountByCountry(w http.ResponseWriter, r *http.Request) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx,"localhost:6002",
     grpc.WithTransportCredentials(insecure.NewCredentials()))
	 if err != nil {
		fmt.Println("Error connecting to warehouse gRPC server:", err)
		return false, err
	 }
	 defer conn.Close()
	 
	grpcClient := updatestockcountforwarehouselocationpb.NewWarehouseClient(conn)

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
  
	warehouseLocationStructure, err := redis.GetHash[[]models.WarehouseStructure]("warehouseLocationsStructure", countryIdStr)
	if err != nil {
	  return false, err
	}
  
	for _, warehouse := range stockCount.Warehouses {
		warehouseStructure, _ := helper.FindByWhere[models.WarehouseStructure, string](warehouseLocationStructure, func(w models.WarehouseStructure) string { return w.Name }, warehouse.Name) 
  
		for _, location := range warehouse.Locations {
		  _, locationIndex := helper.FindByWhere[models.WarehouseLocationEntry, int](&warehouseStructure.Locations, func(l models.WarehouseLocationEntry) int { return l.LocationId }, location.LocationId)
		  warehouseLocationId := warehouseStructure.WarehouseLocationIds[locationIndex]
		  warehouseStructure.Locations = append(warehouseStructure.Locations, models.WarehouseLocationEntry{LocationName: location.LocationName, LocationId: location.LocationId})


		  stockCountMap, _ := structpb.NewStruct(map[string]interface{}{
			"stockCount": stockCount.Warehouses[locationIndex].Sku[location.LocationName],
		  })
          res, err := grpcClient.UpdateStockcountForWarehouselocation(ctx, &updatestockcountforwarehouselocationpb.StockCountUpdateRequest{
			WarehouseId: int64(warehouseStructure.ID),
			WarehouseLocationId: int64(warehouseLocationId),
			StockCount: stockCountMap,
		  })
		  if err != nil {
			return false, err
		  }
		  if !res.Updated {
			return false, fmt.Errorf("failed to update stock count for warehouse location %d", warehouseLocationId)
		  } 		
	}
}
    if err := redis.DeleteHash("stockcount", countryIdStr); err != nil {
    	fmt.Printf("Error %v deleting stockcount for countryId: %s", err, countryIdStr)
    }
	return true, nil
  }