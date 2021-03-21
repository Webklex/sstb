package app

import (
	"../api/poloniex"
	"../utils/config"
	"../utils/values"
	"./notifier"
	"github.com/adshao/go-binance/v2"
	"os"
	"sync"
	"time"
)

type Build struct {
	Number  string `json:"number"`
	Version string `json:"version"`
}

type Config struct {
	*config.Config

	Timezone string `json:"timezone"`
	JobDir   string `json:"job-dir"`

	Provider []*Provider          `json:"provider"`
	Notifier []*notifier.Notifier `json:"notifier"`

	Build     Build    `json:"-"`
	LogOutput *os.File `json:"-"`
}

type Job struct {
	*config.Config

	Id string `json:"-"`

	Symbol    string `json:"symbol"`
	Primary   string `json:"primary"`
	Secondary string `json:"-"`
	OrderDir  string `json:"order-dir"`

	Volume      values.Float `json:"volume,string"`
	Step        values.Float `json:"step,string"`
	BuyStep     values.Float `json:"buy-step,string"`
	SellStep    values.Float `json:"sell-step,string"`
	Fee         values.Float `json:"fee,string"`
	Enabled     bool         `json:"enabled"`
	Alert       *Alert       `json:"alerts"`
	ProviderId  string       `json:"provider"`
	NotifierIds []string     `json:"notifier"`
	Provider    *Provider    `json:"-"`

	stepSize *values.Float `json:"-"`

	orders  map[int64]*Order         `json:"-"`
	balance map[string]*values.Float `json:"-"`

	lastOperation time.Time  `json:"-"`
	mx            sync.Mutex `json:"-"`

	PoloniexClient *poloniex.Config     `json:"-"`
	BinanceClient  *binance.Client      `json:"-"`
	Notifier       []*notifier.Notifier `json:"-"`
}

type Alert struct {
	Buy     bool  `json:"buy"`
	Sell    bool  `json:"sell"`
	Idle    int   `json:"idle"`
	Summary []int `json:"summary"`
}

// https://github.com/binance/binance-spot-api-docs/blob/master/user-data-stream.md

type BinanceEvent struct {
	EventType       string                  `json:"e"`        // "executionReport",        // Event type
	EventTime       int64                   `json:"E"`        // 1499405658658,            // Event time
	Symbol          string                  `json:"s"`        // "ETHBTC",                 // Symbol
	ClientOrderId   string                  `json:"c"`        // "mUvoqJxFIILMdfAW5iGSOW", // Client order ID
	Side            binance.SideType        `json:"S"`        // "BUY",                    // Side
	OrderType       binance.OrderType       `json:"o"`        // "LIMIT",                  // Order type
	TimeInForce     binance.TimeInForceType `json:"f"`        // "GTC",                    // Time in force
	Quantity        values.Float            `json:"q"`        // "1.00000000",             // Order quantity
	Price           values.Float            `json:"p,string"` // "0.10264410",             // Order price
	StopPrice       values.Float            `json:"P,string"` // "0.00000000",             // Stop price
	IcebergQuantity values.Float            `json:"F,string"` // "0.00000000",             // Iceberg quantity
	//`json:"g"`                      // -1,                       // OrderListId
	OriginalClientOrderId string `json:"C"` // null,                     // Original client order ID; This is the ID of the order being canceled
	CurrentExecutionType  string `json:"x"` // "NEW",                    // Current execution type
	// https://github.com/binance/binance-spot-api-docs/blob/master/rest-api.md#enum-definitions
	Status                   binance.OrderStatusType `json:"X"`        // "NEW", "FILLED", "CANCELED"                    // Current order status
	RejectReseason           string                  `json:"r"`        // "NONE",                   // Order reject reason; will be an error code.
	OrderId                  int64                   `json:"i"`        // 4293153,                  // Order ID
	LastExecutedQuantity     values.Float            `json:"l,string"` // "0.00000000",             // Last executed quantity
	CumulativeFilledQuantity values.Float            `json:"z,string"` // "0.00000000",             // Cumulative filled quantity
	LastExecutedPrice        values.Float            `json:"L,string"` // "0.00000000",             // Last executed price
	//`json:"n"`                              // "0",                      // Commission amount
	//`json:"N"`                              // null,                     // Commission asset
	TransactionTime int64 `json:"T"` // 1499405658657,            // Transaction time
	//`json:"t"`                              // -1,                       // Trade ID
	Ignore int64 `json:"I"` // 8641984,                  // Ignore
	// `json:"w"` // true,                     // Is the order on the book?
	// `json:"m"` // false,                    // Is this trade the maker side?
	//`json:"M"`                              // false,                    // Ignore
	TimeCreated             int64        `json:"O"` // 1499405658657,            // Order creation time
	CumulativeQuoteQuantity values.Float `json:"Z"` // "0.00000000",             // Cumulative quote asset transacted quantity
	// `json:"Y"`            // "0.00000000",              // Last quote asset transacted quantity (i.e. lastPrice * lastQty)
	QuoteOrderQuantity values.Float `json:"Q"` // "0.00000000"              // Quote Order Qty

	// https://github.com/binance/binance-spot-api-docs/blob/master/user-data-stream.md#account-update
	Balances []*BinanceBalanceEvent `json:"B"` // Balances
}

type BinanceBalanceEvent struct {
	Asset string       `json:"a"`        // "BTC"            // Asset
	Free  values.Float `json:"f,string"` // "100.00000000"	 // Free
}
