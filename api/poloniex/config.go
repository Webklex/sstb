package poloniex

import (
	"../../utils/values"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	Username string `json:"username"`
	Key      string `json:"key"`
	Secret   string `json:"secret"`

	RestEndpoint      string        `json:"rest-endpoint"`
	WebsocketEndpoint string        `json:"wss-endpoint"`
	Keepalive         bool          `json:"keepalive"`
	Timeout           time.Duration `json:"timeout"`

	Header map[string]string
	Socket *Socket

	Currencies map[string]*Currency
	Pairs      map[string]*Pair

	client      *http.Client
	throttle    <-chan time.Time
	httpTimeout time.Duration
	debug       bool
}

type Symbol struct {
	Symbol                 string   `json:"symbol"`
	Status                 string   `json:"status"`
	BaseAsset              string   `json:"baseAsset"`
	BaseAssetPrecision     int64    `json:"baseAssetPrecision"`
	QuoteAsset             string   `json:"quoteAsset"`
	QuotePrecision         int64    `json:"quotePrecision"`
	QuoteAssetPrecision    int64    `json:"quoteAssetPrecision"`
	OrderTypes             []string `json:"orderTypes"`
	IcebergAllowed         bool     `json:"icebergAllowed"`
	OCOAllowed             bool     `json:"ocoAllowed"`
	IsSpotTradingAllowed   bool     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed bool     `json:"isMarginTradingAllowed"`
	Permissions            []string `json:"permissions"`
}

type Currency struct {
	Id             int          `json:"id"`
	Name           string       `json:"name"`
	HumanType      string       `json:"humanType"`
	Blockchain     string       `json:"blockchain"`
	HexColor       string       `json:"hexColor"`
	TxFee          values.Float `json:"txFee"`
	MinConf        int          `json:"minConf"`
	DepositAddress string       `json:"depositAddress"`
	Disabled       int          `json:"disabled"`
	Delisted       int          `json:"delisted"`
	Frozen         int          `json:"frozen"`
	Geofenced      int          `json:"isGeofenced"`
}

type Pair struct {
	Id            int          `json:"id"`
	Last          values.Float `json:"last"`
	LowestAsk     values.Float `json:"lowestAsk"`
	HighestBid    values.Float `json:"highestBid"`
	PercentChange values.Float `json:"percentChange"`
	BaseVolume    values.Float `json:"baseVolume"`
	QuoteVolume   values.Float `json:"quoteVolume"`
	IsFrozen      string       `json:"isFrozen"`
	High24hr      values.Float `json:"high24hr"`
	Low24hr       values.Float `json:"low24hr"`
}

type MarketUpd struct {
	// Seq is constantly increasing number.
	Seq int64
	// Initial indicates, that it's the entire obook snapshot.
	Initial bool
	// Obooks - updates of an order book.
	OrderBooks []OrderBookUpd
	// Trades - new trades.
	Trades []TradeUpd
}

type AccountUpd struct {
	// Seq is constantly increasing number.
	Seq int64
	// Initial indicates, that it's the entire obook snapshot.
	Initial bool
	// Obooks - updates of an order book.
	Orders []OrderBookUpd
	// Trades - new trades.
	PendingOrders []OpenOrder
	NewOrders     []OpenOrder
	// Trades - new trades.
	TradeOrders []TradeOrder
}

type OpType int

type OrderBookUpd struct {
	// Type is either bid or ask.
	Type OpType
	// Price of an asset.
	Price values.Float
	// Size can be zero, if the order was removed.
	Size values.Float
}

type TradeUpd struct {
	// TradeID - unique trade ID.
	TradeID string
	// Type is either buy or sell -> 0 = someone accepted a buy offer and sold their coins
	Type OpType
	// Size is a trade amount.
	Size values.Float
	// Price is an asset price.
	Price values.Float
	// Date is a trade's unix timestamp.
	Date int64
}

// TickerUpd is a ticker update message.
type TickerUpd struct {
	Pair string
	Ticker
}

type Tickers struct {
	Pair map[string]Ticker
}

type Ticker struct {
	ID            int          `json:"id"`
	Last          values.Float `json:"last,string"`
	LowestAsk     values.Float `json:"lowestAsk,string"`
	HighestBid    values.Float `json:"highestBid,string"`
	PercentChange values.Float `json:"percentChange,string"`
	BaseVolume    values.Float `json:"baseVolume,string"`
	QuoteVolume   values.Float `json:"quoteVolume,string"`
	IsFrozen      int          `json:"isFrozen,string"`
	High24Hr      values.Float `json:"high24hr,string"`
	Low24Hr       values.Float `json:"low24hr,string"`
}

type Socket struct {
	Endpoint string
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	write    chan wsCmd
	mx       sync.Mutex
	cv       *sync.Cond
	conn     *websocket.Conn
	err      error
	ws       chan bool
	subs     map[int]wsSub
}

type wsMessage struct {
	channel json.Number
	seq     json.Number
	data    [][]interface{}
}

type wsSub struct {
	handler func(interface{})
	err     chan error
}

type wsCmd struct {
	message interface{}
	result  chan error
}

type WebsocketCommand struct {
	Command string      `json:"command"`
	Channel interface{} `json:"channel"`
}

type WebsocketResponse struct {
	CurrencyPair string `json:"currencyPair"`
}

type TestInterface struct {
	CurrencyPair string `json:"currencyPair"`
}

type TradeHandler func(c *websocket.Conn, event *TestInterface)
type ErrHandler func(c *websocket.Conn, err error)

// WsConfig webservice configuration
// WsHandler handle raw websocket message
type SocketHandler func(c *websocket.Conn, message []byte)

type TradeOrder struct {
	Number          int64            `json:"orderNumber,string"`
	Amount          values.Float     `json:"amount,string"`
	Type            string           `json:"type"`
	ResultingTrades []ResultingTrade `json:"resultingTrades"`
	ErrorMessage    string           `json:"error"`
}

type ResultingTrade struct {
	Amount  values.Float `json:"amount,string"`
	Date    string       `json:"date"`
	Rate    values.Float `json:"rate,string"`
	Total   values.Float `json:"total,string"`
	TradeID string       `json:"tradeID"`
	Type    string       `json:"type"`
}

type OpenOrder struct {
	OrderNumber int64         `json:"orderNumber,string"`
	Symbol      int           `json:"symbol"`
	Type        string        `json:"type"`
	TypeNum     values.String `json:"type_num,string"`
	Rate        values.Float  `json:"rate,string"`
	Amount      values.Float  `json:"amount,string"`
	Total       values.Float  `json:"total,string"`
}

type OrderBook struct {
	Asks     [][]interface{} `json:"asks"`
	Bids     [][]interface{} `json:"bids"`
	IsFrozen int             `json:"isFrozen,string"`
	Error    string          `json:"error"`
}

type Trade struct {
	GlobalTradeID int64        `json:"globalTradeID"`
	TradeID       int64        `json:"tradeID,string"`
	OrderNumber   int64        `json:"orderNumber,string"`
	Date          string       `json:"date"`
	Type          string       `json:"type"`
	Rate          values.Float `json:"rate,string"`
	Amount        values.Float `json:"amount,string"`
	Total         values.Float `json:"total,string"`
	Fee           values.Float `json:"fee,string"`
}
