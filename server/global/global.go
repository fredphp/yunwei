package global

import (
	"fmt"
	"yunwei/config"
	"yunwei/model/system"
	"yunwei/utils"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB     *gorm.DB
	Logger *zap.Logger
)

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.CONFIG.Mysql.Username,
		config.CONFIG.Mysql.Password,
		config.CONFIG.Mysql.Host,
		config.CONFIG.Mysql.Port,
		config.CONFIG.Mysql.Database,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	// 自动迁移表结构
	autoMigrate()

	// 初始化数据
	initData()
}

func autoMigrate() {
	err := DB.AutoMigrate(
		&system.SysUser{},
		&system.SysRole{},
		&system.SysMenu{},
		&system.SysRoleMenu{},
		&system.SysRoleApi{},
	)
	if err != nil {
		panic("数据库迁移失败: " + err.Error())
	}
}

func initData() {
	// 初始化角色
	initRoles()

	// 初始化菜单
	initMenus()

	// 初始化超级管理员
	initAdminUser()
}

func initRoles() {
	var count int64
	DB.Model(&system.SysRole{}).Count(&count)
	if count > 0 {
		return
	}

	roles := []system.SysRole{
		{Name: "超级管理员", Keyword: "admin", Description: "系统超级管理员，拥有所有权限", Status: 1},
		{Name: "运维人员", Keyword: "operator", Description: "运维人员，拥有服务器管理权限", Status: 1},
		{Name: "普通用户", Keyword: "user", Description: "普通用户，只有查看权限", Status: 1},
	}

	for _, role := range roles {
		DB.Create(&role)
	}
}

func initMenus() {
	var count int64
	DB.Model(&system.SysMenu{}).Count(&count)
	if count > 0 {
		return
	}

	menus := []system.SysMenu{
		// 一级菜单
		{ParentID: 0, Title: "仪表盘", Name: "Dashboard", Path: "/dashboard", Component: "views/dashboard/index", Icon: "Odometer", Sort: 1, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "服务器管理", Name: "Servers", Path: "/servers", Component: "Layout", Icon: "Monitor", Sort: 2, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "Kubernetes", Name: "Kubernetes", Path: "/kubernetes", Component: "Layout", Icon: "Grid", Sort: 3, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "灰度发布", Name: "Canary", Path: "/canary", Component: "Layout", Icon: "Promotion", Sort: 4, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "负载均衡", Name: "LoadBalancer", Path: "/loadbalancer", Component: "Layout", Icon: "Connection", Sort: 5, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "证书管理", Name: "Certificate", Path: "/certificate", Component: "Layout", Icon: "DocumentChecked", Sort: 6, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "CDN管理", Name: "CDN", Path: "/cdn", Component: "Layout", Icon: "Position", Sort: 7, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "智能部署", Name: "Deploy", Path: "/deploy", Component: "Layout", Icon: "Upload", Sort: 8, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "任务调度", Name: "Scheduler", Path: "/scheduler", Component: "Layout", Icon: "Timer", Sort: 9, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "Agent管理", Name: "Agents", Path: "/agents", Component: "Layout", Icon: "Cpu", Sort: 10, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "高可用", Name: "HA", Path: "/ha", Component: "Layout", Icon: "CircleCheck", Sort: 11, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "灾备备份", Name: "Backup", Path: "/backup", Component: "Layout", Icon: "Files", Sort: 12, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "成本控制", Name: "Cost", Path: "/cost", Component: "Layout", Icon: "Coin", Sort: 13, Status: 1, Hidden: 0},
		{ParentID: 0, Title: "系统管理", Name: "System", Path: "/system", Component: "Layout", Icon: "Setting", Sort: 14, Status: 1, Hidden: 0},

		// 服务器管理子菜单
		{ParentID: 2, Title: "服务器列表", Name: "ServerList", Path: "/servers/list", Component: "views/servers/index", Icon: "List", Sort: 1, Status: 1, Hidden: 0},
		{ParentID: 2, Title: "告警管理", Name: "Alerts", Path: "/servers/alerts", Component: "views/alerts/index", Icon: "Bell", Sort: 2, Status: 1, Hidden: 0},

		// Kubernetes子菜单
		{ParentID: 3, Title: "集群管理", Name: "Clusters", Path: "/kubernetes/clusters", Component: "views/kubernetes/index", Icon: "Cluster", Sort: 1, Status: 1, Hidden: 0},

		// 系统管理子菜单
		{ParentID: 14, Title: "用户管理", Name: "UserManage", Path: "/system/user", Component: "views/system/user/index", Icon: "User", Sort: 1, Status: 1, Hidden: 0},
		{ParentID: 14, Title: "角色管理", Name: "RoleManage", Path: "/system/role", Component: "views/system/role/index", Icon: "UserFilled", Sort: 2, Status: 1, Hidden: 0},
		{ParentID: 14, Title: "菜单管理", Name: "MenuManage", Path: "/system/menu", Component: "views/system/menu/index", Icon: "Menu", Sort: 3, Status: 1, Hidden: 0},
	}

	for _, menu := range menus {
		DB.Create(&menu)
	}
}

func initAdminUser() {
	var count int64
	DB.Model(&system.SysUser{}).Count(&count)
	if count > 0 {
		return
	}

	// 获取超级管理员角色
	var adminRole system.SysRole
	DB.Where("keyword = ?", "admin").First(&adminRole)

	// 创建超级管理员
	admin := system.SysUser{
		Username: "admin",
		Password: utils.MD5("admin123"),
		NickName: "超级管理员",
		Email:    "admin@example.com",
		Status:   1,
		RoleID:   adminRole.ID,
	}

	DB.Create(&admin)
	fmt.Println("\n===========================================")
	fmt.Println("  超级管理员账号创建成功!")
	fmt.Println("  用户名: admin")
	fmt.Println("  密码: admin123")
	fmt.Println("===========================================")
}

func InitLogger() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		panic("日志初始化失败: " + err.Error())
	}
}
