package helper

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	models "stock_automation_backend_go/services/api-gateway/api-routes/types"
	"stock_automation_backend_go/shared/env"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func WriteJson(w http.ResponseWriter, statusCode int, result any, outererr error) {

	var response models.APIResponseStruct
	var errString *string

	if outererr != nil {
		msg := outererr.Error()
		errString = &msg
		response = models.APIResponseStruct{StatusCode: statusCode, Response: nil, Error: errString}
	} else {
		response = models.APIResponseStruct{StatusCode: statusCode, Response: result, Error: errString}
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	fmt.Println("Reached here")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}

func WriteMicroServiceJson(w http.ResponseWriter, statusCode int, result any, outererr error) {

	var response any
	//var errString *string

	if outererr != nil {
		// msg := outererr.Error()
		// errString = &msg
		// w.Header().Set("Content-Type", "application/text")
		// w.WriteHeader(statusCode)
		// w.Write([]byte(fmt.Sprintf("error %v", outererr.Error())))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)

		w.Write([]byte(outererr.Error()))

		return
	} else {
		response = result
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	//return jsonData
	w.Write(jsonData)
}

func ComparePassword(password, hashedPassword []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func CreateToken(payload struct {
	ID                           int64
	UserName, DisplayName, Email string
}, tokenType string) (string, error) {
	secret := []byte(env.GetEnv[string](env.EnvKeys.SECRET_KEY))
	exp := time.Now().Add(24 * time.Hour)

	if tokenType != "A" {
		secret = []byte(env.GetEnv[string](env.EnvKeys.REFRESH_SECRET_KEY))
		exp = time.Now().Add(7 * 24 * time.Hour)
	}

	claims := jwt.MapClaims{}
	claims["id"] = payload.ID
	claims["userName"] = payload.UserName
	claims["DisplayName"] = payload.DisplayName
	claims["Email"] = payload.Email
	claims["exp"] = exp.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)

}

func FindByWhere[T any, K comparable](slice *[]T, selector func(T) K, want K) (*T, int) {
	for i, v := range *slice {
		if selector(v) == want {
			return &v, i
		}
	}
	return nil, -1
}

func IsTimeInPast(hours int64, minutes int64, seconds int64, days int64) bool {
	now := time.Now()
	targetTime := time.Date(now.Year(), now.Month(), now.Day()+int(days), int(hours), int(minutes), int(seconds), 0, now.Location())
	return targetTime.Before(now)
}

func JobTimeSetter(hours int64, minutes int64, seconds int64, days int64) (int64, int64, int64, int64) {
	seconds += env.GetEnv[int64](env.EnvKeys.JOB_TIME_SECOND_FREQUENCY)
	minutes += int64(math.Floor(float64(seconds) / 60))
	seconds %= 60

	minutes += env.GetEnv[int64](env.EnvKeys.JOB_TIME_MINUTE_FREQUENCY)
	hours += int64(math.Floor(float64(minutes) / 60))
	minutes %= 60

	hours += env.GetEnv[int64](env.EnvKeys.JOB_TIME_HOUR_FREQUENCY)
	days += int64(math.Floor(float64(hours) / 24))
	hours %= 24
	
	return hours, minutes, seconds, days
}

func Job(hours int64, minutes int64, seconds int64, days int64) bool {
	//wg := sync.WaitGroup{}
	now := time.Now()
	targetTime := time.Date(now.Year(), now.Month(), now.Day()+int(days), int(hours), int(minutes), int(seconds), 0, now.Location())

	//wg.Add(1)

	time.Sleep(targetTime.Sub(now))

	//wg.Done()

	return true
}

var (
	JOB_TIME_HOUR   int64 = env.GetEnv[int64]("JOB_TIME_HOUR")
	JOB_TIME_MINUTE int64 = env.GetEnv[int64]("JOB_TIME_MINUTE")
	JOB_TIME_SECOND int64 = env.GetEnv[int64]("JOB_TIME_SECOND")
	JOB_TIME_DAY    int64 = env.GetEnv[int64]("JOB_TIME_DAY")
)

func JobTimeEmit() (int64, int64, int64, int64, map[string]time.Time) {

	for IsTimeInPast(JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY) {
		JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY = JobTimeSetter(JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY)
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

	//socketio.Broadcast("jobScheduledEvent", )
    return JOB_TIME_HOUR, JOB_TIME_MINUTE, JOB_TIME_SECOND, JOB_TIME_DAY, map[string]time.Time{"scheduledForTime": targetTime}
}