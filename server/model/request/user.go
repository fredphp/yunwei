package request

// Register 注册请求
type Register struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	NickName string `json:"nickName"`
}

// Login 登录请求
type Login struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Captcha   string `json:"captcha"`
	CaptchaId string `json:"captchaId"`
}

// UserInfo 用户信息请求
type UserInfo struct {
	ID uint `uri:"id" binding:"required"`
}

// UserList 用户列表请求
type UserList struct {
	Page     int    `form:"page" binding:"required"`
	PageSize int    `form:"pageSize" binding:"required"`
	Username string `form:"username"`
	Status   *int   `form:"status"`
}

// UserCreate 创建用户请求
type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	NickName string `json:"nickName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleId   uint   `json:"roleId"`
}

// UserUpdate 更新用户请求
type UserUpdate struct {
	ID       uint   `json:"id" binding:"required"`
	NickName string `json:"nickName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleId   uint   `json:"roleId"`
	Status   *int   `json:"status"`
}

// SetUserPassword 设置用户密码
type SetUserPassword struct {
	ID          uint   `json:"id" binding:"required"`
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// PageRequest 分页请求
type PageRequest struct {
	Page     int `form:"page" binding:"required"`
	PageSize int `form:"pageSize" binding:"required"`
}

// MenuCreate 创建菜单请求
type MenuCreate struct {
	ParentID  uint   `json:"parentId"`
	Title     string `json:"title" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	Sort      int    `json:"sort"`
	Hidden    int    `json:"hidden"`
}

// MenuUpdate 更新菜单请求
type MenuUpdate struct {
	ID        uint   `json:"id" binding:"required"`
	ParentID  uint   `json:"parentId"`
	Title     string `json:"title" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	Sort      int    `json:"sort"`
	Hidden    int    `json:"hidden"`
	Status    *int   `json:"status"`
}

// RoleCreate 创建角色请求
type RoleCreate struct {
	Name        string `json:"name" binding:"required"`
	Keyword     string `json:"keyword" binding:"required"`
	Description string `json:"description"`
	MenuIds     []uint `json:"menuIds"`
	ApiIds      []uint `json:"apiIds"`
}

// RoleUpdate 更新角色请求
type RoleUpdate struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Keyword     string `json:"keyword" binding:"required"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
	MenuIds     []uint `json:"menuIds"`
	ApiIds      []uint `json:"apiIds"`
}

// ApiCreate 创建API请求
type ApiCreate struct {
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Group       string `json:"group"`
	Description string `json:"description"`
}
