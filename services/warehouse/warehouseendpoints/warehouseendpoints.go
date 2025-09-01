package warehouseendpoints

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/helper"
	"strconv"
	"strings"
	"time"
	"warehouse/types/models"
	"warehouse/types/responsemodels"

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

func GetColumnMappings(w http.ResponseWriter, r *http.Request) ([]models.ColumnMapping, error) {
	DB := database.GetDB()
	ctx := r.Context()

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("warehouse name not provided")
	}
	warehouseName := parts[len(parts)-1]

	query := `SELECT sc.id AS standard_column_id,
	       wcm.id AS warehouse_column_mapping_id,
	       sc.standard_name AS standard_name,
	       wcm.excel_name AS excel_name
	FROM warehouse.warehouse_column_mappings wcm
	INNER JOIN warehouse.standard_columns sc ON wcm.standard_column_id = sc.id
	INNER JOIN warehouse.warehouses w ON wcm.warehouse_id = w.id
	WHERE w.name = $1 AND wcm.is_deleted = false
	ORDER BY sc.id`

	rows, err := DB.QueryContext(ctx, query, warehouseName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []models.ColumnMapping
	for rows.Next() {
		var cm models.ColumnMapping
		if err := rows.Scan(&cm.StandardColumnId, &cm.WarehouseColumnMappingId, &cm.StandardName, &cm.ExcelName); err != nil {
			return nil, err
		}
		mappings = append(mappings, cm)
	}

	return mappings, nil
}

func GetWarehouseLocationsStructure(w http.ResponseWriter, r *http.Request) ([]models.WarehouseStructure, error) {
	DB := database.GetDB()
	ctx := r.Context()

	query := `SELECT wl.id AS warehouse_location_id,
	       wl.country_id AS country_id,
	       c.name AS country_name,
	       w.id AS warehouse_id,
	       l.id AS location_id,
	       wl.name AS warehouse_location_name,
	       l.name AS location_name,
	       w.name AS warehouse_name
	FROM warehouse.warehouse_locations wl
	INNER JOIN warehouse.locations l ON wl.location_id = l.id
	INNER JOIN warehouse.warehouses w ON wl.warehouse_id = w.id
	INNER JOIN warehouse.countries c ON wl.country_id = c.id
	WHERE wl.is_deleted = false
	ORDER BY w.id, wl.id`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type row struct {
		WarehouseLocationId   int
		CountryId             int
		CountryName           string
		WarehouseId           int
		LocationId            int
		WarehouseLocationName string
		LocationName          string
		WarehouseName         string
	}

	var data []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.WarehouseLocationId, &r.CountryId, &r.CountryName, &r.WarehouseId, &r.LocationId, &r.WarehouseLocationName, &r.LocationName, &r.WarehouseName); err != nil {
			return nil, err
		}
		data = append(data, r)
	}

	warehouseMap := map[int]*models.WarehouseStructure{}

	for _, d := range data {
		ws, ok := warehouseMap[d.WarehouseId]
		if !ok {
			ws = &models.WarehouseStructure{
				Name:                 d.WarehouseName,
				ID:                   d.WarehouseId,
				Countries:            []string{},
				Locations:            []models.WarehouseLocationEntry{},
				CountryIds:           []int{},
				WarehouseLocationIds: []int{},
				CountryLocations:     map[string][]string{},
			}
			warehouseMap[d.WarehouseId] = ws
		}

		foundCountry := false
		for _, c := range ws.Countries {
			if c == d.CountryName {
				foundCountry = true
				break
			}
		}
		if !foundCountry {
			ws.Countries = append(ws.Countries, d.CountryName)
		}

		foundLoc := false
		for _, loc := range ws.Locations {
			if loc.LocationName == d.LocationName {
				foundLoc = true
				break
			}
		}
		if !foundLoc {
			ws.Locations = append(ws.Locations, models.WarehouseLocationEntry{LocationName: d.LocationName, LocationId: d.WarehouseLocationId})
		}

		contains := func(arr []int, v int) bool {
			for _, x := range arr {
				if x == v {
					return true
				}
			}
			return false
		}
		if !contains(ws.WarehouseLocationIds, d.WarehouseLocationId) {
			ws.WarehouseLocationIds = append(ws.WarehouseLocationIds, d.WarehouseLocationId)
		}
		if !contains(ws.CountryIds, d.CountryId) {
			ws.CountryIds = append(ws.CountryIds, d.CountryId)
		}

		if _, ok := ws.CountryLocations[d.CountryName]; !ok {
			ws.CountryLocations[d.CountryName] = []string{}
		}
		present := false
		for _, ln := range ws.CountryLocations[d.CountryName] {
			if ln == d.LocationName {
				present = true
				break
			}
		}
		if !present {
			ws.CountryLocations[d.CountryName] = append(ws.CountryLocations[d.CountryName], d.LocationName)
		}
	}

	var result []models.WarehouseStructure
	for _, ws := range warehouseMap {
		result = append(result, *ws)
	}

	return result, nil
}

type Country struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetCountries(w http.ResponseWriter, r *http.Request) ([]Country, error) {
	DB := database.GetDB()
	ctx := r.Context()

	rows, err := DB.QueryContext(ctx, `SELECT id, name FROM warehouse.countries ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var c Country
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}
	return countries, nil
}

func GetStockCountFromHistory(w http.ResponseWriter, r *http.Request) (json.RawMessage, error) {
	DB := database.GetDB()
	ctx := r.Context()

	fmt.Println(strings.Trim(r.URL.Path, "/"))
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("countryId and historyDate required")
	}
	countryIdStr := parts[len(parts)-2]
	historyDate := parts[len(parts)-1]

	countryId, err := strconv.Atoi(countryIdStr)
	if err != nil {
		return nil, fmt.Errorf("invalid country id")
	}

	const query = `
	SELECT stock_count FROM stockkeepingunit.stock_count_history
			 WHERE country_id = $1 AND created_at::date = to_date($2, 'YYYY-MM-DD')
			 ORDER BY created_at DESC, id DESC LIMIT 1`

	var payload json.RawMessage
	if err := DB.QueryRowContext(ctx, query, countryId, historyDate).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return json.RawMessage([]byte(`{}`)), nil
		}
		return nil, err
	}
	return payload, nil
}

func AddStockCountHistoryForCountry(w http.ResponseWriter, r *http.Request) (map[string]any, error) {
	DB := database.GetDB()
	ctx := r.Context()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("country id required")
	}
	countryIdStr := parts[len(parts)-1]
	if _, err := strconv.Atoi(countryIdStr); err != nil {
		return nil, fmt.Errorf("invalid country id")
	}

	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	query := `INSERT INTO public.stock_count_history (stock_count, country_id)
	          VALUES ($1, $2) RETURNING created_at`
	var createdAt time.Time
	if err := DB.QueryRowContext(ctx, query, body, countryIdStr).Scan(&createdAt); err != nil {
		return nil, err
	}
	return map[string]any{"createdAt": createdAt}, nil
}

func GetStockCountByWarehouseCountries(w http.ResponseWriter, r *http.Request) ([]map[string]any, error) {
	type reqBody struct {
		Countries []int `json:"countries"`
		Poll      bool  `json:"poll"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	DB := database.GetDB()
	ctx := r.Context()

	out := []map[string]any{}
	for _, countryID := range body.Countries {
		rows, err := DB.QueryContext(ctx, `SELECT 
			wl.id AS warehouse_location_id,
			c.name AS country_name,
			w.id AS warehouse_id,
			l.id AS location_id,
			wl.name AS warehouse_location_name,
			l.name AS location_name,
			w.name AS warehouse_name,
			wl.skus_count
		FROM warehouse.warehouse_locations wl
		INNER JOIN warehouse.locations l ON wl.location_id = l.id
		INNER JOIN warehouse.warehouses w ON wl.warehouse_id = w.id
		INNER JOIN warehouse.warehouse_countries wc ON wl.warehouse_id = wc.warehouse_id AND wl.country_id = wc.country_id
		INNER JOIN warehouse.countries c ON wl.country_id = c.id
		WHERE wc.country_id = $1 AND COALESCE(wl.is_deleted, false) = false`, countryID)
		if err != nil {
			return nil, err
		}

		type row struct {
			WarehouseLocationId   int
			CountryName           string
			WarehouseId           int
			LocationId            int
			WarehouseLocationName string
			LocationName          string
			WarehouseName         string
			SkusCountRaw          json.RawMessage
		}

		var data []row
		for rows.Next() {
			var rrow row
			if err := rows.Scan(&rrow.WarehouseLocationId, &rrow.CountryName, &rrow.WarehouseId, &rrow.LocationId, &rrow.WarehouseLocationName, &rrow.LocationName, &rrow.WarehouseName, &rrow.SkusCountRaw); err != nil {
				rows.Close()
				return nil, err
			}
			data = append(data, rrow)
		}
		rows.Close()

		type stockWarehouse struct {
			ID        int                             `json:"id"`
			Name      string                          `json:"name"`
			Locations []models.WarehouseLocationEntry `json:"locations"`
			Sku       map[string]models.Sku           `json:"sku"`
		}
		warehouses := []stockWarehouse{}

		for _, d := range data {
			var sku models.Sku
			if err := json.Unmarshal(d.SkusCountRaw, &sku); err != nil {
				return nil, err
			}
			foundIdx := -1
			for idx, wh := range warehouses {
				if wh.Name == d.WarehouseName {
					foundIdx = idx
					break
				}
			}
			if foundIdx == -1 {
				warehouses = append(warehouses, stockWarehouse{ID: d.WarehouseId, Name: d.WarehouseName, Locations: []models.WarehouseLocationEntry{}, Sku: map[string]models.Sku{}})
				foundIdx = len(warehouses) - 1
			}
			exists := false
			for _, loc := range warehouses[foundIdx].Locations {
				if loc.LocationId == d.WarehouseLocationId {
					exists = true
					break
				}
			}
			if !exists {
				warehouses[foundIdx].Locations = append(warehouses[foundIdx].Locations, models.WarehouseLocationEntry{LocationName: d.LocationName, LocationId: d.WarehouseLocationId})
			}
			warehouses[foundIdx].Sku[d.LocationName] = sku
		}

		out = append(out, map[string]any{
			"countryId":   countryID,
			"countryName": func() string { if len(data) > 0 { return data[0].CountryName } ; return "" }(),
			"warehouses":  warehouses,
		})
	}

	return out, nil
}

func UpdateStockCountForWarehouseLocation(w http.ResponseWriter, r *http.Request) (bool, error) {
    DB := database.GetDB()
    ctx := context.Background()

    // Expect path: /updateStockCountForWarehouseLocation/{warehouseId}/{warehouseLocationId}
    parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
    if len(parts) < 3 {
        return false, fmt.Errorf("warehouseId and warehouseLocationId required")
    }
    wlId := parts[len(parts)-1]
    whId := parts[len(parts)-2]

    var bodyParser struct {
        StockCount json.RawMessage `json:"stockCount"`
    }
    if err := json.NewDecoder(r.Body).Decode(&bodyParser); err != nil {
        return false, err
    }
    defer r.Body.Close()

    _, err := DB.ExecContext(ctx, `UPDATE warehouse.warehouse_locations
        SET skus_count = $1
        WHERE warehouse_id = $2 AND id = $3`, &bodyParser , whId, wlId)
    if err != nil {
        return false, err
    }
    return true, nil
}

func UpdateWarehouseColumnMapping(w http.ResponseWriter, r *http.Request) (bool, error) {
	DB := database.GetDB()
	ctx := r.Context()

	var body struct {
		StandardColumnId int    `json:"standardColumnId"`
		WarehouseId      int    `json:"warehouseId"`
		ExcelName        string `json:"excelName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return false, err
	}
	defer r.Body.Close()

	_, err := DB.ExecContext(ctx, `UPDATE warehouse.warehouse_column_mappings
		SET excel_name = $1 WHERE warehouse_id = $2 AND standard_column_id = $3`, body.ExcelName, body.WarehouseId, body.StandardColumnId)
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteWarehouseColumnMapping(w http.ResponseWriter, r *http.Request) (bool, error) {
	DB := database.GetDB()
	ctx := r.Context()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		return false, fmt.Errorf("warehouseId and standardColumnId required")
	}
	stdStr := parts[len(parts)-1]
	whStr := parts[len(parts)-2]
	if _, err := strconv.Atoi(stdStr); err != nil {
		return false, fmt.Errorf("invalid standardColumnId")
	}
	if _, err := strconv.Atoi(whStr); err != nil {
		return false, fmt.Errorf("invalid warehouseId")
	}

	_, err := DB.ExecContext(ctx, `UPDATE warehouse.warehouse_column_mappings
		SET is_deleted = true, deleted_at = NOW()
		WHERE warehouse_id = $1 AND standard_column_id = $2`, whStr, stdStr)
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetStockCountData(w http.ResponseWriter, r *http.Request) (responsemodels.StockCountByWarehouseCountries, error) {
	DB := database.GetDB()
	ctx := r.Context()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if(len(parts) < 2) {
		return responsemodels.StockCountByWarehouseCountries{}, fmt.Errorf("require countryId")
	}

	countryIdStr := parts[len(parts)-1]
	countryId, err := strconv.Atoi(countryIdStr)
	if err != nil {
		return responsemodels.StockCountByWarehouseCountries{}, fmt.Errorf("invalid countryId")
	}

	rows, err := DB.QueryContext(ctx, `SELECT
    wl.id,   c.name AS countryName,    w.id AS warehouseId, l.id AS locationId,   wl.name AS warehouseLocationName,
    l.name AS locationName,   w.name AS warehouseName,  wl.skus_count AS skusCount
FROM warehouse.warehouse_locations wl
INNER JOIN warehouse.locations l
    ON wl.location_id = l.id
INNER JOIN warehouse.warehouses w
    ON wl.warehouse_id = w.id
INNER JOIN warehouse.warehouse_countries wc
    ON wl.warehouse_id = wc.warehouse_id
   AND wl.country_id = wc.country_id
INNER JOIN warehouse.countries c
    ON wl.country_id = c.id
WHERE wc.country_id = $1
  AND wl.is_deleted = false;`, countryId)

	if err != nil {
		return responsemodels.StockCountByWarehouseCountries{}, err
	}

	defer rows.Close()


	type row struct {
		WarehouseLocationId int
		WarehouseId int
		CountryName string
		WarehouseName string
		LocationName string
		SkusCount models.Sku
		WarehouseLocationName string
		LocationId int
	}

	var skusCountRaw json.RawMessage
	var responseRows []row

	for rows.Next() {
	  var responseRow row
	  if err := rows.Scan(&responseRow.WarehouseLocationId, &responseRow.CountryName, &responseRow.WarehouseId,  &responseRow.LocationId, 
		&responseRow.WarehouseLocationName, 
		&responseRow.LocationName, &responseRow.WarehouseName,  &skusCountRaw); err != nil {
	  	return responsemodels.StockCountByWarehouseCountries{}, err
	  }
	  if err := json.Unmarshal(skusCountRaw, &responseRow.SkusCount); err != nil {
		fmt.Println("Error in Unmarshalling skusCountRaw", err)
		continue
	  }

	  responseRows = append(responseRows, responseRow)
   }

   countryNameRow := DB.QueryRowContext(ctx, "SELECT name FROM warehouse.countries WHERE id = $1", countryId)

   var countryName string

   if err := countryNameRow.Scan(&countryName); err != nil {
	return responsemodels.StockCountByWarehouseCountries{}, err
   }
   fmt.Println("Why are you countryName", countryName, len(responseRows), responseRows[0].WarehouseName)


   warehouses := make([]responsemodels.Warehouse, 0)
   for _, responseRow := range responseRows {
      _, idx := helper.FindByWhere(warehouses,
	  func(r responsemodels.Warehouse) string { return r.Name }, responseRow.WarehouseName)
	 
	   if idx == -1 {
		newWarehouse := responsemodels.Warehouse{
			ID: responseRow.WarehouseId,
			Name: responseRow.WarehouseName,
			Locations: make([]models.WarehouseLocationEntry, 0),
			Sku: map[string]models.Sku{},
		}
		warehouses = append(warehouses, newWarehouse)
		idx = len(warehouses) - 1
	  }

	  _, idx2  := helper.FindByWhere(warehouses[idx].Locations,
		func(r models.WarehouseLocationEntry) int { return r.LocationId }, responseRow.LocationId)
      if idx2 == -1 {
		warehouses[idx].Locations = append(warehouses[idx].Locations, models.WarehouseLocationEntry{LocationName: responseRow.LocationName, LocationId: responseRow.LocationId})
		warehouses[idx].Sku[responseRow.LocationName] = responseRow.SkusCount
	  }
	 
   }

   return responsemodels.StockCountByWarehouseCountries{
	CountryId: countryId,
	CountryName: countryName,
	Warehouses: warehouses,
   }, nil

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
// 		UpdateStockCountForWarehouseLocation(map[string]interface{}{"warehouseId": warehouseStructure.ID, "warehouseLocationId": warehouseLocationId}, 
// 		&stockCount.Warehouses[locationIndex].Sku[location.LocationName])
// 	  }
//   }
// }

