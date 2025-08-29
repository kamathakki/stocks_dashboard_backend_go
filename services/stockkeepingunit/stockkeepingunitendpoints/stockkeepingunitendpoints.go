package stockkeepingunitendpoints

import (
	"net/http"
	"stock_automation_backend_go/database"
	"stockkeepingunit/types/models"

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
