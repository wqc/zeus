package main

import (
	"sync"

	"github.com/zeusship/zeus/eye/adaptor"
	"github.com/zeusship/zeus/eye/config"
	"github.com/zeusship/zeus/eye/flag"
	"github.com/zeusship/zeus/util/log"
)

func main() {
	flag.Parse()

	if flag.Help {
		flag.Useage()
		return
	}

	log.InitConsoleLog()

	log.Info("start parse config file: \"%s\"", flag.ConfigPath)
	cfg, err := config.Parse(flag.ConfigPath)
	if err != nil {
		log.Error("parse config file, err: %s", err)
		return
	}

	log.Info("set log, path: \"%s\", size: %d, level: %s",
		cfg.LogCfg.Path, cfg.LogCfg.RotateSize, cfg.LogCfg.Level)
	log.Initlog(cfg.LogCfg.Path, cfg.LogCfg.Level, cfg.LogCfg.RotateSize)

	binance := adaptor.NewBinanceAdaptor(cfg.BinanceCfg)
	binance.Init()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		binance.Run()
	}()

	wg.Wait()
}
