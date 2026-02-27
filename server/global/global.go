package global

import (
        "fmt"
        "yunwei/config"
        "yunwei/model/agent"
        "yunwei/model/ha"
        "yunwei/model/kubernetes"
        "yunwei/model/scheduler"
        "yunwei/model/system"
        "yunwei/model/tenant"
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
        // 系统管理表
        if err := DB.AutoMigrate(
                &system.SysUser{},
                &system.SysRole{},
                &system.SysMenu{},
                &system.SysApi{},
                &system.SysRoleApi{},
                &system.SysRoleMenu{},
        ); err != nil {
                fmt.Println("系统表迁移警告: " + err.Error())
        }

        // Kubernetes 表
        if err := DB.AutoMigrate(
                &kubernetes.Cluster{},
                &kubernetes.ScaleEvent{},
                &kubernetes.HPAConfig{},
                &kubernetes.DeploymentStatus{},
        ); err != nil {
                fmt.Println("K8s表迁移警告: " + err.Error())
        }

        // Agent 表
        if err := DB.AutoMigrate(
                &agent.Agent{},
                &agent.AgentVersion{},
                &agent.AgentConfig{},
                &agent.AgentMetric{},
                &agent.AgentUpgradeTask{},
                &agent.AgentHeartbeatRecord{},
                &agent.AgentRecoverRecord{},
                &agent.GrayReleaseStrategy{},
        ); err != nil {
                fmt.Println("Agent表迁移警告: " + err.Error())
        }

        // 调度器表
        if err := DB.AutoMigrate(
                &scheduler.Task{},
                &scheduler.CronJob{},
                &scheduler.CronExecution{},
                &scheduler.TaskEvent{},
                &scheduler.TaskQueue{},
                &scheduler.TaskExecution{},
                &scheduler.TaskBatch{},
                &scheduler.TaskTemplate{},
                &scheduler.TaskLog{},
        ); err != nil {
                fmt.Println("调度器表迁移警告: " + err.Error())
        }

        // 租户表
        if err := DB.AutoMigrate(
                &tenant.Tenant{},
                &tenant.TenantQuota{},
                &tenant.TenantUser{},
                &tenant.TenantRole{},
                &tenant.TenantInvitation{},
                &tenant.TenantResourceUsage{},
                &tenant.TenantBilling{},
                &tenant.TenantAuditLog{},
        ); err != nil {
                fmt.Println("租户表迁移警告: " + err.Error())
        }

        // 高可用表
        if err := DB.AutoMigrate(
                &ha.ClusterNode{},
                &ha.DistributedLock{},
                &ha.LeaderElection{},
                &ha.HAClusterConfig{},
                &ha.FailoverRecord{},
                &ha.HASession{},
                &ha.ClusterEvent{},
                &ha.NodeMetric{},
        ); err != nil {
                fmt.Println("HA表迁移警告: " + err.Error())
        }

        // 创建唯一索引
        createUniqueIndexes()
}

// createUniqueIndexes 创建唯一索引
func createUniqueIndexes() {
        // 菜单表：同一父菜单下名称唯一
        DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_name_parent ON sys_menus(name, parent_id)")
        
        // API表：路径和方法唯一
        DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_api_path_method ON sys_apis(path, method)")
        
        // 角色-API关联：角色和API唯一
        DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_api ON sys_role_apis(role_id, api_id)")
        
        // 角色-菜单关联：角色和菜单唯一
        DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_menu ON sys_role_menus(role_id, menu_id)")
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
                {ParentID: 0, Title: "告警中心", Name: "Alerts", Path: "/alerts", Component: "views/alerts/index", Icon: "Bell", Sort: 4, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "灰度发布", Name: "Canary", Path: "/canary", Component: "Layout", Icon: "Promotion", Sort: 5, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "负载均衡", Name: "LoadBalancer", Path: "/loadbalancer", Component: "Layout", Icon: "Connection", Sort: 6, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "证书管理", Name: "Certificate", Path: "/certificate", Component: "Layout", Icon: "DocumentChecked", Sort: 7, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "CDN管理", Name: "CDN", Path: "/cdn", Component: "Layout", Icon: "Position", Sort: 8, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "智能部署", Name: "Deploy", Path: "/deploy", Component: "Layout", Icon: "Upload", Sort: 9, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "任务调度", Name: "Scheduler", Path: "/scheduler", Component: "Layout", Icon: "Timer", Sort: 10, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "Agent管理", Name: "Agents", Path: "/agents", Component: "Layout", Icon: "Cpu", Sort: 11, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "高可用", Name: "HA", Path: "/ha", Component: "Layout", Icon: "CircleCheck", Sort: 12, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "灾备备份", Name: "Backup", Path: "/backup", Component: "Layout", Icon: "Files", Sort: 13, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "成本控制", Name: "Cost", Path: "/cost", Component: "Layout", Icon: "Coin", Sort: 14, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "租户管理", Name: "Tenant", Path: "/tenant", Component: "views/tenant/index", Icon: "OfficeBuilding", Sort: 15, Status: 1, Hidden: 0},
                {ParentID: 0, Title: "系统管理", Name: "System", Path: "/system", Component: "Layout", Icon: "Setting", Sort: 99, Status: 1, Hidden: 0},

                // 服务器管理子菜单
                {ParentID: 2, Title: "服务器列表", Name: "ServerList", Path: "/servers/list", Component: "views/servers/index", Icon: "List", Sort: 1, Status: 1, Hidden: 0},

                // Kubernetes子菜单
                {ParentID: 3, Title: "集群管理", Name: "Clusters", Path: "/kubernetes/clusters", Component: "views/kubernetes/index", Icon: "Cluster", Sort: 1, Status: 1, Hidden: 0},

                // 系统管理子菜单
                {ParentID: 16, Title: "用户管理", Name: "UserManage", Path: "/system/user", Component: "views/system/user/index", Icon: "User", Sort: 1, Status: 1, Hidden: 0},
                {ParentID: 16, Title: "角色管理", Name: "RoleManage", Path: "/system/role", Component: "views/system/role/index", Icon: "UserFilled", Sort: 2, Status: 1, Hidden: 0},
                {ParentID: 16, Title: "菜单管理", Name: "MenuManage", Path: "/system/menu", Component: "views/system/menu/index", Icon: "Menu", Sort: 3, Status: 1, Hidden: 0},
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
