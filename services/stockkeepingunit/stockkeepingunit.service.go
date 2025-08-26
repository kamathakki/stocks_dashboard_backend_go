package stockkeepingunit

import (
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/services/stockkeepingunit/types/models"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func GetStockKeepingUnits(w http.ResponseWriter, r *http.Request) {
	DB := database.GetDB()
	ctx := r.Context()

	dbResponse, err := DB.QueryContext(ctx, `SELECT id, name, sku_code, excel_names, model_no, weight, carton_weight, fitting_in_carton
                    FROM stock_keeping_units WHERE is_deleted = false ORDER BY id`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	helper.WriteJson(w, http.StatusOK, stockKeepingUnitRows)

}
