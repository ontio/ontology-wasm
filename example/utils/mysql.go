package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var (
	dbInstance *sql.DB
)

func InitMysql() {
	if db, err := sql.Open("mysql", "universe:123456@tcp(localhost:3306)/home"); err != nil {
		panic(err)
	} else {
		if err = db.Ping(); err != nil {
			panic(err)
			fmt.Println("connect mysql fail")
		} else {
			fmt.Println("connected mysql")
		}
		dbInstance = db
	}
}

func GetDB() *sql.DB {
	return dbInstance
}
