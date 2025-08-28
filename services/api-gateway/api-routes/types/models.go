package models

type APIResponseStruct struct {
	StatusCode int     `json:"statusCode"`
	Response   any     `json:"response"`
	Error      *string `json:"error"`
}
