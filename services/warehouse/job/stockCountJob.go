package stockCountJob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/shared/env"
	"strconv"
	"time"
	"warehouse/types/responsemodels"
	"warehouse/warehouseendpoints"

	socketio "stock_automation_backend_go/services/socketio"
)

var (
	JOB_TIME_HOUR   int64 = env.GetEnv[int64]("JOB_TIME_HOUR")
	JOB_TIME_MINUTE int64 = env.GetEnv[int64]("JOB_TIME_MINUTE")
	JOB_TIME_SECOND int64 = env.GetEnv[int64]("JOB_TIME_SECOND")
	JOB_TIME_DAY    int64 = env.GetEnv[int64]("JOB_TIME_DAY")
)

func RunJob() {
	getCountriesReq, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("/warehouse/getCountries"), nil)
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return
	}

	countries, err := warehouseendpoints.GetCountries(nil, getCountriesReq)
	if err != nil {
		fmt.Println("Error getting countries: ", err)
		return
	}

	currSCInCountries := make(map[int]responsemodels.StockCountByWarehouseCountries, len(countries))

	JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY, _ = helper.JobTimeEmit()
	isJobTime := helper.Job(JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY)

	if isJobTime == true {
		for _, country := range countries {
			go func() {
				getStockCountDataReq, err := http.NewRequest(http.MethodGet,
					fmt.Sprintf("/warehouse/getStockCountData/%d", country.ID), nil)

				if err != nil {
					fmt.Println("Error creating request: ", err)
					return
				}

				currSCInCountries[country.ID], err = warehouseendpoints.GetStockCountData(nil, getStockCountDataReq)
				if err != nil {
					fmt.Println("Error getting stock count data: ", err)
					return
				}

				b, _ := json.Marshal(currSCInCountries[country.ID])
				reqAdd, _ := http.NewRequest(http.MethodPost, "/warehouse/addStockCountHistoryForCountry/"+strconv.Itoa(country.ID), bytes.NewReader(b))
				reqAdd.Header.Set("Content-Type", "application/json")

				responseForCountry, err := warehouseendpoints.AddStockCountHistoryForCountry(nil, reqAdd)
				if err != nil {
					fmt.Println("Error adding stock count history for country: ", err)
					return
				}

				executionMetaData := map[string]time.Time{"createdAt": responseForCountry["createdAt"]}
				socketio.Broadcast("jobExecutedEvent-"+strconv.Itoa(country.ID), executionMetaData)

			}()
		}

	}
	RunJob()
}
