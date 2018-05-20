package huobi

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zeusship/zeus/util/log"

	"github.com/gorilla/websocket"
)

type MarketDetailDataResponse struct {
	Market    string        `json:"ch"`
	Timestamp int64         `json:"ts"`
	Data      []*MarketData `json:"data"`
}

type MarketData struct {
	ID          int64   `json:"id"`
	Price       float64 `json:"price"`
	Timestamp   int64   `json:"time"`
	Amount      float64 `json:"amount"`
	Direction   string  `json:"buy"`
	TradeID     int64   `json:"tradeId"`
	TimestampMS int64   `json:"1494495766000"`
}

type MarketDetailSubResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Subbed    string `json:"subbed"`
	Timestamp int64  `json:"ts"`
}

type MarketDetailResponse struct {
	MarketDetailSubResponse
	MarketDetailDataResponse
}

const (
	huobiGET  = 1
	huobiPOST = 2
)

type HuobiService struct {
	endpoint string
	apiID    string
	apiKey   string
}

func NewHuobiService(wsendpoint string, apiid, apikey string) *HuobiService {
	hs := new(HuobiService)
	hs.endpoint = wsendpoint
	hs.apiID = apiid
	hs.apiKey = apikey

	return hs
}

func (hs *HuobiService) Prepare(req *http.Request) error {
	Timestamp := time.Now().UTC().Format("2006-01-02T15:04:05")
	SignatureVersion := "2"
	SignatureMethod := "HmacSHA256"
	AccessKeyId := hs.apiID

	values := url.Values{}

	if req.Method == http.MethodGet {
		uv := req.URL.Query()
		for k, vs := range uv {
			if k == "Timestamp" ||
				k == "SignatureVersion" ||
				k == "SignatureMethod" ||
				k == "AccessKeyId" {
				continue
			}

			for _, s := range vs {
				values.Add(k, s)
			}
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else if req.Method == http.MethodPost {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Accept-Language", "zh-cn")

	values.Set("Timestamp", Timestamp)
	values.Set("SignatureVersion", SignatureVersion)
	values.Set("SignatureMethod", SignatureMethod)
	values.Set("AccessKeyId", AccessKeyId)

	qu := values.Encode()

	ps := req.Method + "\n" +
		strings.ToLower(req.Host) + "\n" +
		req.RequestURI + "\n" + qu

	mac := hmac.New(sha256.New, []byte(hs.apiKey))
	mac.Write([]byte(ps))
	sq := mac.Sum(nil)
	sbq := base64.RawURLEncoding.EncodeToString(sq)

	req.URL.Query().Add("Signature", sbq)

	return nil
}

func (hs *HuobiService) WebSocketMarketDetail(symbol string) (chan *MarketDetailDataResponse, error) {
	ch := make(chan *MarketDetailDataResponse, 16)
	conn, _, err := websocket.DefaultDialer.Dial(hs.endpoint, nil)
	if err != nil {
		log.Error("huobi websocket dial faild, err: %s, endpoint: %s", err, hs.endpoint)
		return nil, err
	}

	pingHandler := func(appdata string) error {
		err := conn.WriteControl(websocket.PongMessage,
			[]byte(strings.Replace(appdata, "ping", "pong", -1)),
			time.Now().Add(time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		}

		if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}

		return err
	}

	conn.SetPingHandler(pingHandler)

	go func() {
		for {
			t, r, err := conn.NextReader()
			if err != nil {
				log.Error("huobi websocket exited by err: %s", err)
				break
			}
			switch t {
			case websocket.TextMessage:
			case websocket.BinaryMessage:
				gr, err := gzip.NewReader(r)
				if err != nil {
					log.Error("huobi new reader faild, err: %s", err)
					continue
				}

				data, err := ioutil.ReadAll(gr)
				if err != nil {
					log.Error("huobi gzip read faild, err: %s", err)
					continue
				}

				res := new(MarketDetailResponse)
				if err = json.Unmarshal(data, res); err != nil {
					log.Error("houbi api unmarl faild, err: %s", err)
					continue
				}

				if len(res.Data) == 0 {
					continue
				}

				ch <- &res.MarketDetailDataResponse
			}
		}
	}()

	return ch, nil
}
