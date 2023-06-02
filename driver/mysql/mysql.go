package mysql

import (
	"fmt"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Mysql 组件
type Mysql struct {
	Option *Option

	*gorm.DB
}

type Option struct {
	Username     string `flag:"username" default:"" usage:"Mysql username"`
	Password     string `flag:"password" default:"" usage:"Mysql password"`
	Host         string `flag:"host" default:"127.0.0.1" usage:"Mysql host"`
	Port         string `flag:"port" default:"3306" usage:"Mysql port"`
	Databases    string `flag:"databases" default:"" usage:"Which db used by client"`
	Charset      string `flag:"charset" default:"utf8mb4" usage:"Databases charset"`
	Debug        bool   `flag:"debug" default:"false" usage:"debug mode print sql"`
	MaxOpenConns int    `flag:"max_open_conns" default:"1000" usage:"最大连接数"`
	MaxIdleConns int    `flag:"max_idle_conns" default:"50" usage:"最大空闲连接数"`
}

func NewMysql(_ box.Context, opt *Option) (*Mysql, error) {
	m := &Mysql{Option: opt}

	db, err := gorm.Open(mysql.Open(m.dsn()))
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(opt.MaxOpenConns)
	sqlDB.SetMaxIdleConns(opt.MaxIdleConns)

	m.DB = db

	if opt.Debug {
		m.DB = m.DB.Debug()
	}
	return m, nil
}

func (th *Mysql) dsn() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		th.Option.Username, th.Option.Password, th.Option.Host, th.Option.Port, th.Option.Databases, th.Option.Charset)
}
