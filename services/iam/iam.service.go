package iam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/services/iam/types/models"
)

func Login(w http.ResponseWriter, r *http.Request) {
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
	   FROM users u
	   LEFT JOIN roles r on u.role_id = r.id  
	   LEFT JOIN role_permissions rp on r.id = rp.role_id
	   LEFT JOIN permissions p on rp.permission_id = p.id
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(permissionsJSON, &loginResponse.Permissions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(codesJSON, &loginResponse.PermissionCodes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := helper.ComparePassword([]byte(userLoginBody.Password), []byte(loginResponse.Password)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteJson(w, http.StatusOK, models.LoginResponse{User: loginResponse, Token: accessToken, IsLoggedIn: true})

}
