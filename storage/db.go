package storage

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	Alias        string `yaml:"alias"`
	Addr         string `yaml:"addr"`
	User         string `yaml:"user"`
	Database     string `yaml:"database"`
	Password     string `yaml:"password"`
	Port         int    `yaml:"port" default:"3306"`
	MaxIdleConns int    `yaml:"max_idle_connections" default: "10"`
	MaxOpenConns int    `yaml:"max_open_connections" default: "20"`
}

func (cfg *DBConfig) Source() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.User, cfg.Password, cfg.Addr, cfg.Port, cfg.Database)
}

func Open(cfg *DBConfig) error {
	return orm.RegisterDataBase(cfg.Alias, "mysql", cfg.Source(), cfg.MaxIdleConns, cfg.MaxOpenConns)
}
