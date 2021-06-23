package datastore

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/model"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var databaseType = "sqlite3"
var databasePath = "./data/db"
var gdb *gorm.DB

func init() {

	if os.Getenv("DB_TYPE") != "" {
		databaseType = os.Getenv("DB_TYPE")
	}

	if os.Getenv("DB_PATH") != "" {
		databasePath = os.Getenv("DB_PATH")
	}

}

// NewDB creates a new gorm.DB
func NewDB() *gorm.DB {
	var db *gorm.DB
	gormCfg := &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			TablePrefix: "adaptor_",
		},
	}

	if databaseType == "mysql" {
		mySQLConfig := &mysql.Config{
			User:                 config.C.DB.User,
			Passwd:               config.C.DB.Pass,
			Net:                  "tcp",
			Addr:                 fmt.Sprintf("%s:%d", config.C.DB.Host, config.C.DB.Port),
			DBName:               config.C.DB.Name,
			AllowNativePasswords: true,
			ParseTime:            true,
			Loc:                  time.Local,
			Params:               map[string]string{"charset": "utf8"},
			Timeout:              time.Second * 5,
		}

		retry := 10
		for retry > 0 {
			var err error

			db, err = gorm.Open(gmysql.Open(mySQLConfig.FormatDSN()), gormCfg)
			if err != nil {
				logrus.Errorf("open db connection failure %s, will retry", err.Error())
				time.Sleep(time.Second * 3)
				retry--
				continue
			}
			break
		}
	} else {
		os.MkdirAll(databasePath, 0755)
		var err error
		databaseFilePath := path.Join(databasePath, "db.sqlite3")
		db, err = gorm.Open(sqlite.Open(databaseFilePath), gormCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
	gdb = db
	return db
}

//GetGDB -
func GetGDB() *gorm.DB {
	return gdb
}

// AutoMigrate run auto migration for given models
func AutoMigrate(db *gorm.DB) error {
	models := map[string]interface{}{
		"CloudAccessKey": model.CloudAccessKey{},
		"CreateKubernetesTask": model.CreateKubernetesTask{},
		"InitRainbondTask": model.InitRainbondTask{},
		"RKECluster": model.RKECluster{},
		"CustomCluster": model.CustomCluster{},
		"UpdateKubernetesTask": model.UpdateKubernetesTask{},
		"RainbondClusterConfig": model.RainbondClusterConfig{},
		"AppStore": model.AppStore{},
	}

	for name, mod := range models {
		if err := db.AutoMigrate(mod); err != nil {
			return fmt.Errorf("auto migrate %s: %v", name, err)
		}
	}

	return nil
}
