package config

import (
	"github.com/zeusship/zeus/eye/adaptor"
	"github.com/zeusship/zeus/storage"
	"github.com/zeusship/zeus/util/log"

	"github.com/jinzhu/configor"
)

func Parse(path string) (*Config, error) {
	cfg := new(Config)
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	cfg.BinanceCfg.DBConfig = cfg.DBCfg
	cfg.HuobiCfg.DBConfig = cfg.DBCfg

	return cfg, nil
}

type Config struct {
	DBCfg      *storage.DBConfig      `yaml:db`
	LogCfg     *log.LogCfg            `yaml:"log"`
	BinanceCfg *adaptor.BinanceConfig `yaml:"binance"`
	HuobiCfg   *adaptor.HuobiConfig   `yaml:"huobi"`
}
