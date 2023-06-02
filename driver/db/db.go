package db

import (
	"database/sql"
)

type Options struct {
	DSN string `flag:"dsn" default:"" usage:"连接地址"`
}

func NewDB(opt *Options) (*sql.DB, error) {
	db, err := sql.Open("mysql", opt.DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}
