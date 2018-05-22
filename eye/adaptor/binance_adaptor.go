package adaptor

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/zeusship/zeus/storage"
	"github.com/zeusship/zeus/util/log"

	binance "github.com/binance-exchange/go-binance"
	klog "github.com/go-kit/kit/log"
	klevel "github.com/go-kit/kit/log/level"
)

type BinanceConfig struct {
	EndPoint string            `yaml:"endpoint"`
	ApiKey   string            `yaml:"key"`
	ApiID    string            `yaml:"id"`
	Symbols  []string          `yaml:"symbols"`
	DBConfig *storage.DBConfig `yaml:"db"`
}

func NewBinanceAdaptor(cfg *BinanceConfig) *binanceAdaptor {
	a := new(binanceAdaptor)

	a.cfg = cfg

	return a
}

type binanceAdaptor struct {
	cfg        *BinanceConfig
	signer     binance.Signer
	banApi     binance.Service
	waitSymbol chan string
	wg         *sync.WaitGroup
	marketCh   chan *storage.Market
}

func (ba *binanceAdaptor) Init() {
	signer := new(binance.HmacSigner)
	signer.Key = []byte(ba.cfg.ApiKey)
	ba.signer = signer

	// may be need cancel

	//ba.banApi = banApi

	ba.waitSymbol = make(chan string, len(ba.cfg.Symbols)+1)
	for _, symbol := range ba.cfg.Symbols {
		ba.waitSymbol <- symbol
	}

	ba.wg = new(sync.WaitGroup)
	ba.marketCh = make(chan *storage.Market, 512*len(ba.cfg.Symbols)+10)
}

func (ba *binanceAdaptor) Run() {
	for {
		symbol, isExited := <-ba.waitSymbol
		if !isExited {
			log.Info("binance adaptor accept close sign")
			return
		}

		log.Info("accept symbol: %s", symbol)
		ba.wg.Add(1)
		go ba.runSymbol(symbol)
		ba.wg.Wait()
	}
}

func (ba *binanceAdaptor) Close() {
	log.Info("binance adptor exiting")
	close(ba.waitSymbol)
	ba.wg.Wait()
	log.Info("binance adptor exited")
}

func (ba *binanceAdaptor) runSymbol(symbol string) {
	defer ba.wg.Done()
	ctx := context.Background()
	logger := klog.NewLogfmtLogger(klog.NewSyncWriter(os.Stderr))
	logger = klevel.NewFilter(logger, klevel.AllowAll())
	logger = klog.With(logger, "time", klog.DefaultTimestampUTC, "caller", klog.DefaultCaller)

	banApi := binance.NewAPIService(
		ba.cfg.EndPoint,
		ba.cfg.ApiID,
		ba.signer,
		logger,
		ctx,
	)

	tch, done, err := banApi.DepthWebsocket(binance.DepthWebsocketRequest{Symbol: symbol})
	if err != nil {
		log.Error("new binance ws API faild, err: %s", err)
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
			for _, b := range t.Bids {
				m := new(storage.Market)
				m.Type = storage.BidType
				m.Price = b.Price
				m.Quantity = b.Quantity
				m.Symbol = t.Symbol
				m.Exchange = "binance"
				m.Timestamp = t.Time.UnixNano() / int64(time.Millisecond)
				ba.marketCh <- m
			}

			for _, a := range t.Asks {
				m := new(storage.Market)
				m.Type = storage.AskType
				m.Price = a.Price
				m.Quantity = a.Quantity
				m.Symbol = t.Symbol
				m.Exchange = "binance"
				m.Timestamp = t.Time.UnixNano() / int64(time.Millisecond)
				ba.marketCh <- m
			}

		case <-done:
			log.Info("endpoint finish send data")
			time.Sleep(time.Second * 5)
			break
		}
	}

	log.Info("endpoint: %s, symbol: %d, exited after recv trade: %d ", ba.cfg.EndPoint, symbol, count)

	ba.waitSymbol <- symbol
}

func (ba *binanceAdaptor) writeBinanceTrade() {
	orm, err := storage.Orm(ba.cfg.DBConfig.Alias)
	if err != nil {
		log.Error("binance new orm faild, err: %s", err)
		return
	}

	for {
		m := <-ba.marketCh
		if _, err := orm.Insert(m); err != nil {
			log.Error("binance inster trade faild, err: %s, trade: %s", err, m)
		}
	}
}
