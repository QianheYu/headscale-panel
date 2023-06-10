package common

import (
	"errors"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/util"
)

// InitData init postgres data
func InitData() {
	// To initialise data or not
	if config.Conf.System.InitData {
		preparingData()
	}
	// Adjustment of the availability of some debugging interfaces and menus according to the operating mode
	hidden := 2
	if config.GetMode() >= config.MULTI {
		hidden = 1
	}
	if err := DB.Model(&model.Menu{}).Where("path in (?)", []string{"acl"}).Update("hidden", hidden).Error; err != nil {
		log.Log.Errorf("write hidden menu data to database error：%v", err)
	}

	hidden = 1
	if config.Conf.System.Mode == "debug" {
		hidden = 2
	}
	if err := DB.Model(&model.Menu{}).Where("path in (?)", []string{"api", "menu"}).Update("hidden", hidden).Error; err != nil {
		log.Log.Errorf("write hidden menu data to database error: %v", err)
	}
}

func preparingData() {
	// 1. write roles
	newRoles := make([]*model.Role, 0)
	roles := []*model.Role{
		{
			Name:    "administrator",
			Keyword: "admin",
			Desc:    new(string),
			Home:    "/dashboard",
			Sort:    1,
			Status:  1,
			Creator: "System",
		},
		{
			Name:    "manager",
			Keyword: "manager",
			Desc:    new(string),
			Home:    "/console",
			Sort:    2,
			Status:  1,
			Creator: "System",
		},
		{
			Name:    "user",
			Keyword: "user",
			Desc:    new(string),
			Home:    "/console",
			Sort:    3,
			Status:  1,
			Creator: "System",
		},
	}

	for _, role := range roles {
		err := DB.First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newRoles = append(newRoles, role)
		}
	}

	if len(newRoles) > 0 {
		err := DB.Create(&newRoles).Error
		if err != nil {
			log.Log.Errorf("write role data to database error：%v", err)
		}
	}

	// 2. write menus
	newMenus := make([]model.Menu, 0)
	//messageStr := "message"
	componentStr := "component"
	dashboardStr := "dashboard"
	dashboardRedir := "/dashboard/index"
	//messagePathStr := "/message/system"
	systemDashboardStr := "/system/dashboard"
	//userStr := "user"
	//peoplesStr := "peoples"
	//treeTableStr := "tree-table"
	//treeStr := "tree"
	exampleStr := "example"
	logOperationStr := "/log/operation-log"
	//documentationStr := "documentation"
	consoleStr := "/console/machines"

	menus := []model.Menu{
		{
			Name:       "Dashboard",
			Title:      "Dashboard",
			Icon:       &dashboardStr,
			Path:       "/dashboard",
			Component:  "Layout",
			Redirect:   &dashboardRedir,
			Breadcrumb: 2,
			Sort:       1,
			ParentId:   0,
			Roles:      roles[:1],
			Creator:    "System",
		},
		{
			Name:      "Dashboard",
			Title:     "Dashboard",
			Icon:      &dashboardStr,
			Path:      "index",
			Component: "/dashboard/index",
			Sort:      2,
			ParentId:  1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:       "Console",
			Title:      "Console",
			Icon:       &componentStr,
			Path:       "/console",
			Component:  "Layout",
			Redirect:   &consoleStr,
			AlwaysShow: 1,
			Sort:       10,
			ParentId:   0,
			Roles:      roles[:3],
			Creator:    "System",
		},
		{
			Name:      "MachinesManage",
			Title:     "Machines",
			Path:      "machines",
			Component: "/console/machines/index",
			Sort:      11,
			ParentId:  3,
			Roles:     roles[:2],
			Creator:   "System",
		},
		{
			Name:      "MachinesCommon",
			Title:     "Machines",
			Path:      "machines",
			Component: "/console/machines/index-common",
			Sort:      12,
			ParentId:  3,
			Roles:     roles[2:],
			Creator:   "System",
		},
		{
			Name:      "Routes",
			Title:     "Routes",
			Path:      "routes",
			Component: "/console/routes/index",
			Sort:      13,
			ParentId:  3,
			Roles:     roles[:2],
			Creator:   "System",
		},
		{
			Name:      "ACL",
			Title:     "Access Control",
			Path:      "acl",
			Component: "/console/acl/index",
			Sort:      14,
			ParentId:  3,
			Roles:     roles[:2],
			//Roles:   roles[2:3],
			Creator: "System",
		},
		{
			Name:      "Setting",
			Title:     "Setting",
			Path:      "setting",
			Component: "/console/setting/index",
			Sort:      15,
			ParentId:  3,
			Roles:     roles,
			Creator:   "System",
		},
		//{
		//	Name:      "Message Center",
		//	Title:     "消息中心",
		//	Icon:      &messageStr,
		//	Path:      "/message",
		//	Component: "Layout",
		//	Redirect:  &messagePathStr,
		//	Sort:      20,
		//	ParentId:  0,
		//	Roles:     roles[:2],
		//	Creator:   "System",
		//},
		//{
		//	Name:      "System Message",
		//	Title:     "系统消息",
		//	Path:      "index",
		//	Component: "/message/index",
		//	Sort:      21,
		//	ParentId:  8,
		//	Roles:     roles[:2],
		//	Creator:   "System",
		//},
		{
			Name:       "System",
			Title:      "System",
			Icon:       &componentStr,
			Path:       "/system",
			Component:  "Layout",
			Redirect:   &systemDashboardStr,
			AlwaysShow: 1,
			Sort:       30,
			ParentId:   0,
			Roles:      roles[:2],
			Creator:    "System",
		},
		{
			Name:  "User",
			Title: "User",
			//Icon:      &userStr,
			Path:      "user",
			Component: "/system/user/index",
			Sort:      31,
			ParentId:  9,
			Roles:     roles[:2],
			Creator:   "System",
		},
		{
			Name:  "Role",
			Title: "Role",
			//Icon:      &peoplesStr,
			Path:      "role",
			Component: "/system/role/index",
			Sort:      32,
			ParentId:  9,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:  "Menu",
			Title: "Menu",
			//Icon:      &treeTableStr,
			Path:      "menu",
			Component: "/system/menu/index",
			Sort:      33,
			ParentId:  9,
			Hidden:    1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:  "Api",
			Title: "Api",
			//Icon:      &treeStr,
			Path:      "api",
			Component: "/system/api/index",
			Sort:      34,
			ParentId:  9,
			Hidden:    1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Headscale",
			Title:     "Headscale",
			Path:      "headscale",
			Component: "/system/headscale/index",
			Sort:      35,
			ParentId:  9,
			Roles:     roles[:1],
			Creator:   "System",
		},
		//{
		//	Name:      "Setting",
		//	Title:     "Setting",
		//	Path:      "setting",
		//	Component: "/system/setting/index",
		//	Sort:      36,
		//	ParentId:  9,
		//	Roles:     roles[:1],
		//	Creator:   "System",
		//},
		{
			Name:       "Log",
			Title:      "Log",
			Icon:       &exampleStr,
			Path:       "/log",
			Component:  "Layout",
			Redirect:   &logOperationStr,
			AlwaysShow: 1,
			Sort:       40,
			ParentId:   0,
			Roles:      roles[:1],
			Creator:    "System",
		},
		{
			Name:  "OperationLog",
			Title: "Operation Log",
			//Icon:      &documentationStr,
			Path:      "operation-log",
			Component: "/log/operation-log/index",
			Sort:      41,
			ParentId:  15,
			Roles:     roles[:1],
			Creator:   "System",
		},
	}
	for _, menu := range menus {
		err := DB.First(&menu).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newMenus = append(newMenus, menu)
		}
	}
	if len(newMenus) > 0 {
		err := DB.Create(&newMenus).Error
		if err != nil {
			log.Log.Errorf("write menu to database error：%v", err)
		}
	}

	// 3. write users
	newUsers := make([]model.User, 0)
	users := []model.User{
		{
			Name:     "admin",
			Password: util.GenPasswd("123456"),
			//Mobile:       "18888888888",
			Email:        "admin@example.com",
			Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
			Nickname:     "",
			Introduction: "",
			Status:       1,
			Creator:      "System",
			Roles:        roles[:1],
		},
		//{
		//	Username:     "faker",
		//	Password:     util.GenPasswd("123456"),
		//	Mobile:       "19999999999",
		//	Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		//	Nickname:     new(string),
		//	Introduction: new(string),
		//	Status:       1,
		//	Creator:      "System",
		//	Roles:        roles[2:3],
		//},
		//{
		//	Username:     "nike",
		//	Password:     util.GenPasswd("123456"),
		//	Mobile:       "13333333333",
		//	Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		//	Nickname:     new(string),
		//	Introduction: new(string),
		//	Status:       1,
		//	Creator:      "System",
		//	Roles:        roles[3:4],
		//},
		//{
		//	Username:     "bob",
		//	Password:     util.GenPasswd("123456"),
		//	Mobile:       "15555555555",
		//	Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		//	Nickname:     new(string),
		//	Introduction: new(string),
		//	Status:       1,
		//	Creator:      "System",
		//	Roles:        roles[3:4],
		//},
	}

	for _, user := range users {
		err := DB.First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newUsers = append(newUsers, user)
		}
	}

	if len(newUsers) > 0 {
		if err := DB.Create(&newUsers).Error; err != nil {
			log.Log.Errorf("write user data to database error：%v", err)
		}
	}

	// 4. write headscale setting
	newHeadscaleSetting := &model.Headscale{
		GRPCServerAddr: "localhost:50443",
		ApiKey:         "",
		Insecure:       false,
	}

	if err := DB.First(&newHeadscaleSetting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		if err = DB.Create(&newHeadscaleSetting).Error; err != nil {
			log.Log.Errorf("write headscale setting to database error：%v", err)
		}
	}

	// 5. write apis
	apis := []model.Api{
		{
			Method:   "POST",
			Path:     "/base/login",
			Category: "base",
			Desc:     "User login",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/base/logout",
			Category: "base",
			Desc:     "User logout",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/base/refreshToken",
			Category: "base",
			Desc:     "Refresh JWT token",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/info",
			Category: "user",
			Desc:     "Get current logged-in user information",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/user/list",
			Category: "user",
			Desc:     "Get user list",
			Creator:  "System",
		},
		{
			Method:   "PUT",
			Path:     "/user/changePwd",
			Category: "user",
			Desc:     "Update user login password",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/create",
			Category: "user",
			Desc:     "Create user",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/user/update/:userId",
			Category: "user",
			Desc:     "Update user",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/user/delete/batch",
			Category: "user",
			Desc:     "Batch delete users",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/role/list",
			Category: "role",
			Desc:     "Get role list",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/role/create",
			Category: "role",
			Desc:     "Create role",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/role/update/:roleId",
			Category: "role",
			Desc:     "Update role",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/role/menus/get/:roleId",
			Category: "role",
			Desc:     "Get role' permission menus",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/role/menus/update/:roleId",
			Category: "role",
			Desc:     "Update role's permission menus",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/role/apis/get/:roleId",
			Category: "role",
			Desc:     "Get role's permission APIs",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/role/apis/update/:roleId",
			Category: "role",
			Desc:     "Update role's permission APIs",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/role/delete/batch",
			Category: "role",
			Desc:     "Batch delete roles",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/menu/list",
			Category: "menu",
			Desc:     "Get menu list",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/menu/tree",
			Category: "menu",
			Desc:     "Get menu tree",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/menu/create",
			Category: "menu",
			Desc:     "Create menu",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/menu/update/:menuId",
			Category: "menu",
			Desc:     "Update menu",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/menu/delete/batch",
			Category: "menu",
			Desc:     "Batch delete menus",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/menu/access/list/:userId",
			Category: "menu",
			Desc:     "Get user's accessible menu list",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/menu/access/tree/:userId",
			Category: "menu",
			Desc:     "Get user's accessible menu tree",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/api/list",
			Category: "api",
			Desc:     "Get API list",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/api/tree",
			Category: "api",
			Desc:     "Get API tree",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/api/create",
			Category: "api",
			Desc:     "Create API",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/api/update/:roleId",
			Category: "api",
			Desc:     "Update API",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/api/delete/batch",
			Category: "api",
			Desc:     "Batch delete APIs",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/system/info",
			Category: "system",
			Desc:     "Get system info",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/system/status",
			Category: "system",
			Desc:     "Get system status",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/system/install",
			Category: "system",
			Desc:     "Install headscale",
			Creator:  "System",
		},
		//{
		//	Method:   "GET",
		//	Path:     "/system/setting",
		//	Category: "setting",
		//	Desc:     "Get setting",
		//	Creator:  "System",
		//},
		//{
		//	Method:   "PATCH",
		//	Path:     "/system/setting",
		//	Category: "setting",
		//	Desc:     "Update setting",
		//	Creator:  "System",
		//},
		{
			Method:   "GET",
			Path:     "/system/headscale",
			Category: "headscale",
			Desc:     "Get headscale config",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/system/headscale",
			Category: "headscale",
			Desc:     "Update headscale config",
			Creator:  "System",
		},
		// 接口预留
		//{
		//	Method:   "POST",
		//	Path:     "/system/headscale/upload/:target",
		//	Category: "headscale",
		//	Desc:     "Upload file",
		//	Creator:  "System",
		//},
		{
			Method:   "GET",
			Path:     "/log/operation/list",
			Category: "log",
			Desc:     "Get operation log list",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/log/operation/delete/batch",
			Category: "log",
			Desc:     "Batch delete operation logs",
			Creator:  "System",
		},
		//{
		//	Method:   "GET",
		//	Path:     "/notice",
		//	Category: "notice",
		//	Desc:     "获取站内通知",
		//	Creator:  "System",
		//},
		//{
		//	Method:   "GET",
		//	Path:     "/message",
		//	Category: "message",
		//	Desc:     "获取系统消息",
		//	Creator:  "System",
		//},
		//{
		//	Method:   "POST",
		//	Path:     "/message",
		//	Category: "message",
		//	Desc:     "标记系统消息",
		//	Creator:  "System",
		//},
		//{
		//	Method:   "DELETE",
		//	Path:     "/message",
		//	Category: "message",
		//	Desc:     "删除系统消息",
		//	Creator:  "System",
		//},
		{
			Method:   "GET",
			Path:     "/console/preauthkey",
			Category: "console",
			Desc:     "Get PreAuthKey list",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/console/preauthkey",
			Category: "console",
			Desc:     "Create PreAuthKey",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/console/preauthkey",
			Category: "console",
			Desc:     "Expire PreAuthKey",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/console/acl",
			Category: "console",
			Desc:     "Get Access Control",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/console/acl",
			Category: "console",
			Desc:     "Save Access Control",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/console/route",
			Category: "console",
			Desc:     "Get machine route",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/console/route",
			Category: "console",
			Desc:     "Switch route",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/console/route",
			Category: "console",
			Desc:     "Delete route",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/console/machine",
			Category: "console",
			Desc:     "Get machine list",
			Creator:  "System",
		},
		{
			Method:   "PUT",
			Path:     "/console/machine",
			Category: "console",
			Desc:     "Move machine",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/console/machine",
			Category: "console",
			Desc:     "Add or Update machine",
			Creator:  "System",
		},
		{
			Method:   "DELETE",
			Path:     "/console/machine",
			Category: "console",
			Desc:     "Delete machine",
			Creator:  "System",
		},
		{
			Method:   "PATCH",
			Path:     "/console/machine",
			Category: "console",
			Desc:     "Set machine tag",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/.well-known/openid-configuration",
			Category: "oidc",
			Desc:     "Get OIDC API",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/oidc/authorize",
			Category: "oidc",
			Desc:     "Authorize",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/oidc/token",
			Category: "oidc",
			Desc:     "Get access token",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/oidc/user_info",
			Category: "oidc",
			Desc:     "Get user info",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/oidc/jwk",
			Category: "oidc",
			Desc:     "Get jwk",
			Creator:  "System",
		},
	}

	// different role has different paths permission
	basePaths := []string{
		"/base/login",
		"/base/logout",
		"/base/refreshToken",
		"/user/info",
		"/menu/access/tree/:userId",
	}
	tailnetPaths := []string{
		"/role/list",
		"/user/list",
		"/user/create",
		"/user/changePwd",
		"/user/update/:userId",
		"/user/delete/batch",
		"/log/operation/list",
		"/log/operation/delete/batch",
		"/message",
		"/notice",
		"/console/preauthkey",
		"/console/acl",
		"/console/routes",
		"/console/route",
		"/console/machine",
		"/oidc/authorize",
	}
	userPaths := []string{
		"/console/preauthkey",
		"/console/acl",
		"/console/routes",
		"/console/route",
		"/console/machine",
		"/oidc/authorize",
	}

	newApi := make([]model.Api, 0)
	newRoleCasbin := make([]model.RoleCasbin, 0)
	for i, api := range apis {
		api.ID = uint(i + 1)
		err := DB.First(&api, api.ID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newApi = append(newApi, api)

			//administrator role has full apis permission
			newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
				Keyword: roles[0].Keyword,
				Path:    api.Path,
				Method:  api.Method,
			})

			// no-developer role has basic apis permission
			if funk.ContainsString(basePaths, api.Path) {
				newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
					Keyword: roles[1].Keyword,
					Path:    api.Path,
					Method:  api.Method,
				})
				newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
					Keyword: roles[2].Keyword,
					Path:    api.Path,
					Method:  api.Method,
				})
			}

			// admin role has apis permission
			//if funk.ContainsString(adminPaths, api.Path) {
			//	newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
			//		Keyword: roles[1].Keyword,
			//		Path:    api.Path,
			//		Method:  api.Method,
			//	})
			//}

			// tailnet role has apis permission
			if funk.ContainsString(tailnetPaths, api.Path) {
				newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
					Keyword: roles[1].Keyword,
					Path:    api.Path,
					Method:  api.Method,
				})
			}

			// user role has apis permission
			if funk.ContainsString(userPaths, api.Path) {
				newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
					Keyword: roles[2].Keyword,
					Path:    api.Path,
					Method:  api.Method,
				})
			}
		}
	}

	if len(newApi) > 0 {
		if err := DB.Create(&newApi).Error; err != nil {
			log.Log.Errorf("write api data to database error：%v", err)
		}
	}

	if len(newRoleCasbin) > 0 {
		rules := make([][]string, 0)
		for _, c := range newRoleCasbin {
			rules = append(rules, []string{
				c.Keyword, c.Path, c.Method,
			})
		}
		isAdd, err := CasbinEnforcer.AddPolicies(rules)
		if !isAdd {
			log.Log.Errorf("write casbin data to database error：%v", err)
		}
	}
}
