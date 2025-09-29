package stockCountJob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/shared/env"
	"strconv"
	"time"
	"warehouse/types/responsemodels"
	"warehouse/warehouseendpoints"

	addstockcounthistorybycountrypb "stockkeepingunit/proto"

	socketio "stock_automation_backend_go/services/socketio"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

				ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
				defer cancel()

				conn, err := grpc.DialContext(ctx, fmt.Sprintf("%v:7001", env.GetEnv[string](env.EnvKeys.SKU_HOST)), grpc.WithTransportCredentials(insecure.NewCredentials()))
				grpcClient := addstockcounthistorybycountrypb.NewStockKeepingUnitClient(conn)

				fmt.Println("Reached here mother 1", currSCInCountries[country.ID].Warehouses);

				responseForCountry, err := grpcClient.AddStockCountHistoryforCountry(ctx, &addstockcounthistorybycountrypb.StockCountAddRequestByCountry{
					CountryId: int64(country.ID),
					CountryName: country.Name,
					Warehouses: func(ws []responsemodels.Warehouse) []*addstockcounthistorybycountrypb.Warehouse {
						out := make([]*addstockcounthistorybycountrypb.Warehouse, 0, len(ws))
						for _, w := range ws {
							pw := &addstockcounthistorybycountrypb.Warehouse{
								Id:        int64(w.ID),
								Name:      w.Name,
								Sku:       make(map[string]*addstockcounthistorybycountrypb.SkuCounts, len(w.Sku)),
								Locations: make([]*addstockcounthistorybycountrypb.Location, 0, len(w.Locations)),
							}
							for locName, s := range w.Sku {
								pw.Sku[locName] = &addstockcounthistorybycountrypb.SkuCounts{
									Gks:    int64(s.GKS),
									Neo:    int64(s.NEO),
									Pro:    int64(s.PRO),
									MmB:    int64(s.MMB),
									MmP:    int64(s.MMP),
									Swap:   int64(s.SWAP),
									M3Mr:   int64(s.M3MR),
									M3Pb:   int64(s.M3PB),
									M3Fifa: int64(s.M3FIFA),
								}
							}
							for _, loc := range w.Locations {
								pw.Locations = append(pw.Locations, &addstockcounthistorybycountrypb.Location{
									LocationId:   int64(loc.LocationId),
									LocationName: loc.LocationName,
								})
							}
							out = append(out, pw)
						}
						return out
					}(currSCInCountries[country.ID].Warehouses),
				})
				if err != nil {
					fmt.Println("Error adding stock count history for country: ", err)
					return
				}

				executionMetaData := map[string]string{"createdAt": responseForCountry.GetCreatedAt()}
				socketio.Broadcast("jobExecutedEvent-"+strconv.Itoa(country.ID), executionMetaData)

			}()
		}

	}
	RunJob()
}
