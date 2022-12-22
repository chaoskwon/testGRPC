package common

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

func ConfigureDatabase() *xorm.Engine {
	uid := "root"
	pwd := os.Getenv("SHARING_PLATFORM_DB_PASSWORD")

	dbConnection := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/sharing?charset=utf8mb4&collation=utf8mb4_unicode_ci", uid, pwd)
	xormDb, err := xorm.NewEngine("mysql", dbConnection)
	if err != nil {
		panic(fmt.Errorf("Database open error: " + err.Error()))
	}

	return xormDb
}
