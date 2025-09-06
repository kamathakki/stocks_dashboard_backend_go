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

func jobTimeEmit() {

	for helper.IsTimeInPast(JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY) {
		JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY = helper.JobTimeSetter(JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY)
	}

	now := time.Now()
	targetTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day()+int(JOB_TIME_DAY),
		int(JOB_TIME_HOUR),
		int(JOB_TIME_MINUTE),
		int(JOB_TIME_SECOND), 0, now.Location())
	fmt.Println("Job Time: ", targetTime)

	io := socketio.GetServer()
	io.BroadcastToNamespace("/", "jobScheduledEvent", map[string]time.Time{"scheduledForTime": targetTime})

}

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

	jobTimeEmit()
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

				io := socketio.GetServer()

				b, _ := json.Marshal(currSCInCountries[country.ID])
				reqAdd, _ := http.NewRequest(http.MethodPost, "/warehouse/addStockCountHistoryForCountry/"+strconv.Itoa(country.ID), bytes.NewReader(b))
				reqAdd.Header.Set("Content-Type", "application/json")

				responseForCountry, err := warehouseendpoints.AddStockCountHistoryForCountry(nil, reqAdd)
				if err != nil {
					fmt.Println("Error adding stock count history for country: ", err)
					return
				}

				executionMetaData := map[string]time.Time{"createdAt": responseForCountry["createdAt"]}
				io.BroadcastToNamespace("/", "jobExecutedEvent-"+strconv.Itoa(country.ID), executionMetaData)

			}()
		}

	}
	RunJob()
}
