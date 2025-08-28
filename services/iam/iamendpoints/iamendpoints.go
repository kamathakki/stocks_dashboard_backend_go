package iamendpoints

import (
	"encoding/json"
	"fmt"
	"iam/types/models"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/helper"
)

func Login(w http.ResponseWriter, r *http.Request) (models.LoginResponse, error) {
	var userLoginBody models.LoginBody
	json.NewDecoder(r.Body).Decode(&userLoginBody)

	DB := database.GetDB()
	ctx := r.Context()

	defer r.Body.Close()

	query := `SELECT u.id, u.first_name, u.last_name, u.user_name, u.password, u.email, u.mobile_no,
	   COALESCE(
	   json_agg(
	     json_build_object(
            'permissionName', p.name,
            'permissionCode', p.code,
            'rolePermissionId', rp.id
		 )  
	   ), '[]'::json) AS permissions,
	   r.name,
	   r.code,
	   COALESCE(
	   json_agg(
	     DISTINCT p.code 
	   ), '[]'::json) AS permissionCodes
	   FROM iam.users u
	   LEFT JOIN iam.roles r on u.role_id = r.id  
	   LEFT JOIN iam.role_permissions rp on r.id = rp.role_id
	   LEFT JOIN iam.permissions p on rp.permission_id = p.id
	   WHERE u.user_name = $1 or u.email = $2
	   GROUP BY
       u.id, u.first_name, u.last_name, u.user_name, u.email, u.mobile_no,
       r.name, r.code;`

	dbResponse := DB.QueryRowContext(ctx, query, userLoginBody.UserName, userLoginBody.Email)

	var loginResponse models.LoginResponseUser
	var permissionsJSON, codesJSON json.RawMessage

	if err := dbResponse.Scan(&loginResponse.ID, &loginResponse.FirstName,
		&loginResponse.LastName, &loginResponse.UserName, &loginResponse.Password, &loginResponse.Email,
		&loginResponse.MobileNo, &permissionsJSON,
		&loginResponse.RoleName, &loginResponse.RoleCode, &codesJSON); err != nil {
		return models.LoginResponse{}, fmt.Errorf("no user found")
	}

	if err := json.Unmarshal(permissionsJSON, &loginResponse.Permissions); err != nil {
		return models.LoginResponse{}, err
	}
	if err := json.Unmarshal(codesJSON, &loginResponse.PermissionCodes); err != nil {
		return models.LoginResponse{}, err
	}

	if err := helper.ComparePassword([]byte(userLoginBody.Password), []byte(loginResponse.Password)); err != nil {
		return models.LoginResponse{}, err
	}
	displayName := fmt.Sprintf("%v %v", loginResponse.FirstName, loginResponse.LastName)

	accessToken, err := helper.CreateToken(struct {
		ID                           int64
		UserName, DisplayName, Email string
	}{
		loginResponse.ID,
		loginResponse.UserName,
		displayName,
		loginResponse.Email,
	}, "A")

	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{User: loginResponse, Token: accessToken, IsLoggedIn: true}, nil

}
