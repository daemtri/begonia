package db

import (
	"github.com/jmoiron/sqlx"
)

type Options struct {
	DriverName string `flag:"driver_name" default:"mysql" usage:"驱动"`
	DSN        string `flag:"dsn" default:"" usage:"连接地址"`
}

type Database struct {
	*sqlx.DB
}

func NewDB(opt *Options) (*Database, error) {
	db, err := sqlx.Open(opt.DriverName, opt.DSN)
	if err != nil {
		return nil, err
	}
	return &Database{
		DB: db,
	}, nil
}
