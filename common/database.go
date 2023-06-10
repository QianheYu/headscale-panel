package common

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/model"
)

// Global database variable
var DB *gorm.DB

// Initialize database
func InitDB() {
	var dialector gorm.Dialector
	switch config.Conf.Database.Driver {
	case "mysql":
		dialector = mysql.Open(config.Conf.Database.Dsn)
	case "postgres":
		dialector = postgres.Open(config.Conf.Database.Dsn)
	default:
		panic("Database driver type error")
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		// Disable foreign key constraints (no real foreign key constraints will be created in mysql when specified)
		DisableForeignKeyConstraintWhenMigrating: true,
		//// Specify table prefix
		//NamingStrategy: schema.NamingStrategy{
		//	TablePrefix: config.Conf.Mysql.TablePrefix + "_",
		//},
	})
	if err != nil {
		log.Log.Panicf("Init Database Error: %v", err)
		panic(fmt.Errorf("Init Database Error: %v", err))
	}

	// Enable database logging
	if config.Conf.Database.LogMode {
		db.Debug()
	}
	// Global DB assignment
	DB = db
	// Automatically migrate table structure
	dbAutoMigrate()
	log.Log.Infof("Init database finished.")
}

// Automatically migrate table structure
func dbAutoMigrate() {
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.Headscale{},
		&model.Api{},
		&model.OperationLog{},
		//&model.Message{},
	); err != nil {
		log.Log.Error(err)
		panic(err)
	}
}
