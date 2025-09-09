package iamendpoints

import (
	"encoding/json"
	"fmt"
	"iam/types/models"
	"net/http"
	"stock_automation_backend_go/database"
	"stock_automation_backend_go/helper"
	"strconv"
	"strings"
	"time"

	"stock_automation_backend_go/database/redis"

	"golang.org/x/crypto/bcrypt"
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

	refreshToken, err := helper.CreateToken(struct {
		ID                           int64
		UserName, DisplayName, Email string
	}{
		loginResponse.ID,
		loginResponse.UserName,
		displayName,
		loginResponse.Email,
	}, "R")

	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{User: loginResponse, Token: accessToken, IsLoggedIn: true, RefreshToken: refreshToken}, nil

}

type RegisterBody struct {
	UserName        string `json:"userName"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	MobileNo        string `json:"mobileNo"`
	RoleId          *int   `json:"roleId"`
}

func Register(w http.ResponseWriter, r *http.Request) (map[string]any, error) {
	var body RegisterBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if body.Password != body.ConfirmPassword {
		return nil, fmt.Errorf("passwords do not match")
	}
	if body.Email == "" || !strings.Contains(body.Email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}
	if body.MobileNo != "" && len(body.MobileNo) != 10 {
		return nil, fmt.Errorf("invalid mobile number")
	}

	DB := database.GetDB()
	ctx := r.Context()

	var exists int
	if err := DB.QueryRowContext(ctx, `SELECT 1 FROM iam.users WHERE user_name = $1 LIMIT 1`, body.UserName).Scan(&exists); err == nil {
		return nil, fmt.Errorf("username is already taken")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO iam.users (first_name, last_name, user_name, email, password, mobile_no, role_id)
              VALUES ($1, $2, $3, $4, $5, $6, $7)
              RETURNING user_name, email, created_at`

	var userName, email string
	var createdAt time.Time
	if err := DB.QueryRowContext(ctx, query, body.FirstName, body.LastName, body.UserName, body.Email, string(hashed), body.MobileNo, body.RoleId).Scan(&userName, &email, &createdAt); err != nil {
		return nil, err
	}

	// Invalidate IAM user cache
	_ = redis.DeleteKey("iam:getUsers")

	return map[string]any{"userName": userName, "email": email, "createdAt": createdAt}, nil
}

type UpdateUserBody struct {
	Email    *string `json:"email"`
	RoleId   *int    `json:"roleId"`
	MobileNo *string `json:"mobileNo"`
}

func UpdateUser(w http.ResponseWriter, r *http.Request) (bool, error) {
	var body UpdateUserBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return false, err
	}
	defer r.Body.Close()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		return false, fmt.Errorf("user id required")
	}
	userIdStr := parts[len(parts)-1]
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return false, fmt.Errorf("invalid user id")
	}

	if body.Email == nil && body.RoleId == nil && body.MobileNo == nil {
		return false, fmt.Errorf("at least one field to update is required")
	}

	DB := database.GetDB()
	ctx := r.Context()

	// Build dynamic update
	setParts := []string{}
	args := []any{}
	idx := 1
	if body.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", idx))
		args = append(args, *body.Email)
		idx++
	}
	if body.RoleId != nil {
		setParts = append(setParts, fmt.Sprintf("role_id = $%d", idx))
		args = append(args, *body.RoleId)
		idx++
	}
	if body.MobileNo != nil {
		setParts = append(setParts, fmt.Sprintf("mobile_no = $%d", idx))
		args = append(args, *body.MobileNo)
		idx++
	}
	args = append(args, userId)

	q := fmt.Sprintf("UPDATE iam.users SET %s WHERE id = $%d", strings.Join(setParts, ", "), idx)
	if _, err := DB.ExecContext(ctx, q, args...); err != nil {
		return false, err
	}

	// Invalidate IAM user cache
	_ = redis.DeleteKey("iam:getUsers")

	return true, nil
}

type IGetUsersResponse struct {
	ID              int64    `json:"id"`
	UserName        string   `json:"userName"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	Email           string   `json:"email"`
	MobileNo        string   `json:"mobileNo"`
	RoleName        *string  `json:"roleName"`
	RoleCode        *string  `json:"roleCode"`
	PermissionCodes []string `json:"permissionCodes"`
	Permissions     []struct {
		PermissionCode   string `json:"permissionCode"`
		PermissionName   string `json:"permissionName"`
		RolePermissionId int    `json:"rolePermissionId"`
	} `json:"permissions"`
}

func GetUsers(w http.ResponseWriter, r *http.Request) ([]IGetUsersResponse, error) {
	DB := database.GetDB()
	ctx := r.Context()

	// cache first
	if cached, err := redis.GetKey[[]IGetUsersResponse]("iam:getUsers"); err == nil && cached != nil {
		return *cached, nil
	}

	query := `SELECT 
        u.id, u.first_name, u.last_name, u.user_name, u.email, u.mobile_no,
        r.name as role_name, r.code as role_code,
        COALESCE(json_agg(json_build_object(
            'permissionName', p.name,
            'permissionCode', p.code,
            'rolePermissionId', rp.id
        ) ORDER BY rp.id) FILTER (WHERE p.id IS NOT NULL), '[]'::json) AS permissions,
        COALESCE(json_agg(p.code ORDER BY p.code) FILTER (WHERE p.id IS NOT NULL), '[]'::json) AS permissionCodes
      FROM iam.users u
      LEFT JOIN iam.roles r ON u.role_id = r.id
      LEFT JOIN iam.role_permissions rp ON r.id = rp.role_id
      LEFT JOIN iam.permissions p ON rp.permission_id = p.id
      WHERE COALESCE(u.is_deleted, false) = false
      GROUP BY u.id, u.first_name, u.last_name, u.user_name, u.email, u.mobile_no, r.name, r.code
      ORDER BY u.id`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []IGetUsersResponse{}
	for rows.Next() {
		var u IGetUsersResponse
		var permsJSON, codesJSON json.RawMessage
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.UserName, &u.Email, &u.MobileNo, &u.RoleName, &u.RoleCode, &permsJSON, &codesJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permsJSON, &u.Permissions); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(codesJSON, &u.PermissionCodes); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	// set cache
	if b, err := json.Marshal(users); err == nil {
		_ = redis.SetKey("iam:getUsers", string(b))
	}

	return users, nil
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func GetRoles(w http.ResponseWriter, r *http.Request) ([]Role, error) {
	DB := database.GetDB()
	ctx := r.Context()

	// cache first
	if cached, err := redis.GetKey[[]Role]("iam:getRoles"); err == nil && cached != nil {
		return *cached, nil
	}

	rows, err := DB.QueryContext(ctx, `SELECT id, name, code FROM iam.roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var rl Role
		if err := rows.Scan(&rl.ID, &rl.Name, &rl.Code); err != nil {
			return nil, err
		}
		roles = append(roles, rl)
	}

	// set cache
	if b, err := json.Marshal(roles); err == nil {
		_ = redis.SetKey("iam:getRoles", string(b))
	}

	return roles, nil
}

func DeleteUser(w http.ResponseWriter, r *http.Request) (bool, error) {
	DB := database.GetDB()
	ctx := r.Context()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) < 2 {
		return false, fmt.Errorf("user id required")
	}
	userIdStr := parts[len(parts)-1]
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return false, fmt.Errorf("invalid user id")
	}

	res, err := DB.ExecContext(ctx, "DELETE from iam.users WHERE id = $1", userId)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, fmt.Errorf("user not found")
	}

	// Invalidate IAM user cache
	_ = redis.DeleteKey("iam:getUsers")

	return true, nil
}
