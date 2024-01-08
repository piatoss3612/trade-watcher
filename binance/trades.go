package binance

type AggregateTrade struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	TradeId   int64  `json:"a"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	FirstId   int64  `json:"f"`
	LastId    int64  `json:"l"`
	TradeTime int64  `json:"T"`
	IsMarket  bool   `json:"m"`
	Ignore    bool   `json:"M"`
}

type Request struct {
	Id     int64    `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type Response struct {
	Result []string `json:"result"`
	Id     int64    `json:"id"`
}
