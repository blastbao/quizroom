package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/thinkboy/log4go"
)

var db *sql.DB

func InitMysql() {
	var err error
	if db, err = getDb(); err != nil{
		log4go.Error("mysql connect error %v", err)
	}
	log4go.Info("mysql connect success: %v", Conf.MysqlAddr)
}

func getDb()(db *sql.DB, err error) {
	db, err = sql.Open("mysql", Conf.MysqlUsername+":"+Conf.MysqlPassword+"@tcp("+Conf.MysqlAddr+")/"+Conf.MysqlDatabase+"?charset=utf8mb4")
	db.SetMaxOpenConns(Conf.MysqlMaxOpenConns)
	db.SetMaxIdleConns(Conf.MysqlMaxIdleConns)
	db.Ping()
	return
}
