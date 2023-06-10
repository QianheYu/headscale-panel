package common

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"headscale-panel/config"
	"headscale-panel/log"
)

// Global CasbinEnforcer
var CasbinEnforcer *casbin.Enforcer

// Initialising the casbin policy manager
func InitCasbinEnforcer() {
	e, err := databaseCasbin()
	if err != nil {
		log.Log.Panicf("Failed to initialise Casbin：%v", err)
		panic(fmt.Sprintf("Failed to initialise Casbin：%v", err))
	}

	CasbinEnforcer = e
	log.Log.Info("Initialization of Casbin complete")
}

func databaseCasbin() (*casbin.Enforcer, error) {
	a, err := gormadapter.NewAdapterByDB(DB)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(config.Conf.Casbin.ModelPath, a)
	if err != nil {
		return nil, err
	}

	err = e.LoadPolicy()
	if err != nil {
		return nil, err
	}
	return e, nil
}
