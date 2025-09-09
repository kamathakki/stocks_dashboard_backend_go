package models

type User struct {
	ID        int64  `json:"id"`
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	MobileNo  string `json:"mobile_no"`
	RoleId    int    `json:"role_id"`
	Email     string `json:"email"`
}

type LoginBody struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginResponseUser struct {
	ID              int64    `json:"id"`
	UserName        string   `json:"userName"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	Password        string   `json:"password"`
	MobileNo        string   `json:"mobile_no"`
	Email           string   `json:"email"`
	RoleName        string   `json:"roleName"`
	RoleCode        string   `json:"roleCode"`
	PermissionCodes []string `json:"permissionCodes"`
	Permissions     []struct {
		RolePermissionId int
		PermissionCode   string
		PermissionName   string
	} `json:"permissions"`
}

type LoginResponse struct {
	User         LoginResponseUser `json:"user"`
	Token        string            `json:"token"`
	RefreshToken string            `json:"refreshToken"`
	IsLoggedIn   bool              `json:"isLoggedIn"`
}
