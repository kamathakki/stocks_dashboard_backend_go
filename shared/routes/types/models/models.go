package models

type APIResponseStruct struct {
	StatusCode int   `json:"statusCode"`
	Response   any   `json:"response"`
	Error      error `json:"error"`
}
