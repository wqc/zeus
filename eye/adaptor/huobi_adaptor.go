package adaptor

import (
	"sync"
	"time"

	"github.com/zeusship/zeus/storage"
	"github.com/zeusship/zeus/util/huobi"
	"github.com/zeusship/zeus/util/log"
)

type HuobiConfig struct {
	EndPoint string            `yaml:"endpoint"`
	ApiKey   string            `yaml:"key"`
	ApiID    string            `yaml:"id"`
	Symbols  []string          `yaml:"symbols"`
	DBConfig *storage.DBConfig `yaml:"db"`
}

func NewHuobiAdaptor(cfg *HuobiConfig) *huobiAdaptor {
	a := new(huobiAdaptor)

	a.cfg = cfg

	return a
}

type huobiAdaptor struct {
	cfg        *HuobiConfig
	huobiApi   *huobi.HuobiService
	waitSymbol chan string
	wg         *sync.WaitGroup
	marketCh   chan *storage.Market
}

func (ha *huobiAdaptor) Init() {
	ha.huobiApi = huobi.NewHuobiService(ha.cfg.EndPoint, ha.cfg.ApiID, ha.cfg.ApiKey)

	ha.waitSymbol = make(chan string, len(ha.cfg.Symbols)+1)
	for _, symbol := range ha.cfg.Symbols {
		ha.waitSymbol <- symbol
	}

	ha.wg = new(sync.WaitGroup)
	ha.marketCh = make(chan *storage.Market, 512*len(ha.cfg.Symbols)+10)
}

func (ha *huobiAdaptor) Run() {
	for {
		symbol, isExited := <-ha.waitSymbol
		if !isExited {
			log.Info("huobi adaptor accept close sign")
			return
		}

		log.Info("accept symbol: %s", symbol)
		ha.wg.Add(1)
		go ha.runSymbol(symbol)
	}
}

func (ha *huobiAdaptor) Close() {
	close(ha.waitSymbol)
	ha.wg.Wait()
}

func (ha *huobiAdaptor) runSymbol(symbol string) {
	defer ha.wg.Done()

	tch, err := ha.huobiApi.WebSocketMarketDetail(symbol)
	if err != nil {
		log.Error("new huobi ws API faild, err: %s", err)
		return
	}

	count := int64(0)

	for {
		select {
		case t, ok := <-tch:
			if !ok {
				// 有可能是限流导致，   这里需要具体看情况是否需要直接重试
				time.Sleep(time.Second * 1)
				break
			}
			count++

			for _, d := range t.Data {
				m := new(storage.Market)
				switch d.Direction {
				case "ask":
					m.Type = storage.AskType
				case "bid":
					m.Type = storage.BidType
				}
				m.Price = d.Price
				m.Quantity = d.Amount
				m.Symbol = symbol
				m.Timestamp = d.TimestampMS
				ha.marketCh <- m
			}
		}
	}

	log.Info("endpoint: %s, symbol: %d, exited after recv trade: %d ", ha.cfg.EndPoint, symbol, count)

	ha.waitSymbol <- symbol
}

func (ha *huobiAdaptor) writeHupbiTrade() {
	orm, err := storage.Orm(ha.cfg.DBConfig.Alias)
	if err != nil {
		log.Error("huobi new orm faild, err: %s", err)
		return
	}

	for {
		m := <-ha.marketCh
		if _, err := orm.Insert(m); err != nil {
			log.Error("huobi inster trade faild, err: %s, trade: %s", err, m)
		}
	}
}
