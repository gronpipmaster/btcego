package btcego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const debug = false

const (
	endpointUrl       = "https://btc-e.com/tapi"
	endpointPublicUrl = "https://btc-e.com/api/2/%s/%s"
)

const (
	OrderAsc  = "ASC"
	OrderDesc = "DESC"
)

const (
	OperationTypeBuy  = "buy"
	OperationTypeSell = "sell"
)

type Auth struct {
	AccessKey, SecretKey string
}

type Btce struct {
	auth  Auth
	nonce int64
}

//See https://btc-e.com/api/documentation
func New(auth Auth) *Btce {
	return &Btce{auth, time.Now().Unix()}
}

// Error encapsulates an error returned by btc-e.com
//
type Error struct {
	Message string
}

func (self *Error) Error() string {
	return fmt.Sprintf("Error message: %s", self.Message)
}

type responseWrapper struct {
	Success  int64           `json:"success"`
	ErrorMsg string          `json:"error"`
	Data     json.RawMessage `json:"return"`
}

func (self *Btce) query(params map[string]string, resp interface{}, usingWrapp bool) error {
	self.nonce = self.nonce + 1
	params["nonce"] = fmt.Sprint(self.nonce)
	sign := NewSign(self.auth, params)
	if debug {
		fmt.Println("[--Params--]")
		fmt.Printf("%#v\n", params)
		fmt.Println("[--Signature--]")
		fmt.Println(sign.signature)
	}
	req, err := http.NewRequest("POST", endpointUrl, bytes.NewBufferString(multimap(params).Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Key", sign.key)
	req.Header.Set("Sign", sign.signature)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", fmt.Sprint(len(multimap(params).Encode())))
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// var rawResponse map[string]interface{}
	// json.NewDecoder(r.Body).Decode(&rawResponse)
	// fmt.Println(rawResponse)

	if usingWrapp {
		var respWrapper responseWrapper
		err = json.NewDecoder(r.Body).Decode(&respWrapper)
		if err != nil {
			return err
		}
		if respWrapper.Success == 0 {
			return buildError(respWrapper.ErrorMsg)
		}
		err = json.Unmarshal(respWrapper.Data, resp)
	} else {
		err = json.NewDecoder(r.Body).Decode(resp)
	}

	return err
}

func (self *Btce) queryWoAuth(pair Pair, action string, resp interface{}) error {
	endpoint, err := url.Parse(fmt.Sprintf(endpointPublicUrl, fmt.Sprint(pair), action))
	if err != nil {
		return err
	}
	r, err := http.Get(endpoint.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(resp)
	return err
}

func multimap(p map[string]string) url.Values {
	q := make(url.Values, len(p))
	for k, v := range p {
		q[k] = []string{v}
	}
	return q
}

func buildError(msg string) error {
	return &Error{msg}
}

func makeParams(action string) map[string]string {
	params := make(map[string]string)
	params["method"] = action
	return params
}

func addOptions(params map[string]string, option interface{}) map[string]string {
	options := make(map[string]interface{})
	rawJsonOption, _ := json.Marshal(option)
	json.Unmarshal([]byte(string(rawJsonOption)), &options)
	for key, val := range options {
		params[key] = fmt.Sprint(val)
	}
	return params
}

type Funds struct {
	Usd float64 `json:"usd"`
	Btc float64 `json:"btc"`
	Ltc float64 `json:"ltc"`
	Nmc float64 `json:"nmc"`
	Rur float64 `json:"rur"`
	Eur float64 `json:"eur"`
	Nvc float64 `json:"nvc"`
	Trc float64 `json:"trc"`
	Ppc float64 `json:"ppc"`
	Ftc float64 `json:"ftc"`
	Xpm float64 `json:"xpm"`
}

type Rights struct {
	Info     int64 `json:"info"`
	Trade    int64 `json:"trade"`
	Withdraw int64 `json:"withdraw"`
}

// Example: btc_usd
type Pair string

type OrderId int64

type GetInfoResponse struct {
	Funds            Funds  `json:"funds"`
	Rights           Rights `json:"rights"`
	TransactionCount int64  `json:"transaction_count"`
	OpenOrders       int64  `json:"open_orders"`
	ServerTime       int64  `json:"server_time"`
}

//It returns the information about the user's current balance, API key privileges,the number of transactions, the number of open orders and the server time. See https://btc-e.com/api/documentation section "getInfo"
func (self *Btce) GetInfo() (*GetInfoResponse, error) {
	params := makeParams("getInfo")
	resp := &GetInfoResponse{}
	if err := self.query(params, resp, true); err != nil {
		return nil, err
	}
	return resp, nil
}

type TransHistoryRequest struct {
	//The ID of the transaction to start displaying with, default: 0
	From int64 `json:"from,omitempty"`
	//The number of transactions for displaying, default: 1000
	Count int64 `json:"count,omitempty"`
	//The ID of the transaction to start displaying with, default: 0
	FromId OrderId `json:"from_id,omitempty"`
	//The ID of the transaction to finish displaying with, default: infinity
	EndId OrderId `json:"end_id,omitempty"`
	//Sorting, default btcego.orderDesc
	Order string `json:"order,omitempty"`
	//When to start displaying?, default: 0
	Since int64 `json:"since,omitempty"`
	//When to finish displaying? default: infinity
	End int64 `json:"end,omitempty"`
}

type TransHistoryResponse []TransOrder

type TransOrder struct {
	Type        int64   `json:"type"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"desc"`
	Status      int64   `json:"status"`
	Created     int64   `json:"timestamp"`
}

//It returns the transactions history. See https://btc-e.com/api/documentation section "TransHistory"
func (self *Btce) TransHistory(option *TransHistoryRequest) (*TransHistoryResponse, error) {
	params := makeParams("TransHistory")
	if option != nil {
		addOptions(params, option)
	}
	respWrapp := make(map[string]json.RawMessage)
	if err := self.query(params, &respWrapp, true); err != nil {
		return nil, err
	}
	resp := TransHistoryResponse{}
	for _, rawResp := range respWrapp {
		order := TransOrder{}
		if err := json.Unmarshal(rawResp, &order); err != nil {
			return nil, err
		}
		resp = append(resp, order)
	}
	return &resp, nil
}

type TradeHistoryRequest struct {
	//The ID of the transaction to start displaying with, default: 0
	From int64 `json:"from,omitempty"`
	//The number of transactions for displaying, default: 1000
	Count int64 `json:"count,omitempty"`
	//The ID of the transaction to start displaying with, default: 0
	FromId OrderId `json:"from_id,omitempty"`
	//The ID of the transaction to finish displaying with, default: infinity
	EndId OrderId `json:"end_id,omitempty"`
	//Sorting, default: btcego.OrderDesc
	Order string `json:"order,omitempty"`
	//When to start displaying?, default: 0
	Since int64 `json:"since,omitempty"`
	//When to finish displaying? default: infinity
	End int64 `json:"end,omitempty"`
	//The pair to show the transactions, default: all pairs
	Pair Pair `json:"pair,omitempty"`
}

type TradeHistoryResponse []TradeOrder

type TradeOrder struct {
	Pair        Pair    `json:"pair"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Rate        float64 `json:"rate"`
	OrderId     OrderId `json:"order_id"`
	IsYourOrder int64   `json:"is_your_order"`
	Created     int64   `json:"timestamp"`
}

//It returns the trade history. See https://btc-e.com/api/documentation section "TradeHistory"
func (self *Btce) TradeHistory(option *TradeHistoryRequest) (*TradeHistoryResponse, error) {
	params := makeParams("TradeHistory")
	if option != nil {
		addOptions(params, option)
	}
	respWrapp := make(map[string]json.RawMessage)
	if err := self.query(params, &respWrapp, true); err != nil {
		return nil, err
	}
	resp := TradeHistoryResponse{}
	for _, rawResp := range respWrapp {
		order := TradeOrder{}
		if err := json.Unmarshal(rawResp, &order); err != nil {
			return nil, err
		}
		resp = append(resp, order)
	}
	return &resp, nil
}

type ActiveOrdersRequest struct {
	//The pair to show the transactions, default: all pairs
	Pair Pair `json:"pair,omitempty"`
}

type ActiveOrdersResponse []ActiveOrder

type ActiveOrder struct {
	Pair    Pair    `json:"pair"`
	Type    string  `json:"type"`
	Amount  float64 `json:"amount"`
	Rate    float64 `json:"rate"`
	Status  int64   `json:"status"`
	Created int64   `json:"timestamp_created"`
}

//Returns your open orders. See https://btc-e.com/api/documentation section "ActiveOrders"
func (self *Btce) ActiveOrders(option *ActiveOrdersRequest) (*ActiveOrdersResponse, error) {
	params := makeParams("ActiveOrders")
	if option != nil {
		addOptions(params, option)
	}
	respWrapp := make(map[string]json.RawMessage)
	if err := self.query(params, &respWrapp, true); err != nil {
		return nil, err
	}
	resp := ActiveOrdersResponse{}
	for _, rawResp := range respWrapp {
		order := ActiveOrder{}
		if err := json.Unmarshal(rawResp, &order); err != nil {
			return nil, err
		}
		resp = append(resp, order)
	}
	return &resp, nil
}

type TradeRequest struct {
	//The pair to show the transactions, required
	Pair Pair `json:"pair"`
	//The transaction type (btcego.OperationTypeBuy or btcego.OperationTypeSell), required
	Type string `json:"type"`
	//The rate to buy/sell, required
	Rate float64 `json:"rate"`
	//The amount which is necessary to buy/sell, required
	Amount float64 `json:"amount"`
}

type TradeResponse struct {
	Received float64 `json:"received"`
	Remains  float64 `json:"remains"`
	OrderId  OrderId `json:"order_id"`
	Funds    Funds   `json:"funds"`
}

//Trading is done according to this method. See https://btc-e.com/api/documentation section Trade
func (self *Btce) Trade(option *TradeRequest) (*TradeResponse, error) {
	params := makeParams("Trade")
	if option != nil {
		addOptions(params, option)
	} else {
		return nil, &Error{"TradeRequest is nil."}
	}
	resp := &TradeResponse{}
	if err := self.query(params, resp, true); err != nil {
		return nil, err
	}
	return resp, nil
}

type CancelOrderResponse struct {
	OrderId OrderId `json:"order_id"`
	Funds   Funds   `json:"funds"`
}

//Cancellation of the order. See https://btc-e.com/api/documentation section CancelOrder
func (self *Btce) CancelOrder(orderId OrderId) (*CancelOrderResponse, error) {
	params := makeParams("CancelOrder")
	params["order_id"] = fmt.Sprint(orderId)
	resp := &CancelOrderResponse{}
	if err := self.query(params, resp, true); err != nil {
		return nil, err
	}
	return resp, nil
}

type FeeResponse struct {
	Trade float64 `json:"trade"`
}

//Get service fee, see example https://btc-e.com/api/2/btc_usd/fee
func (self *Btce) GetFee(pair Pair) (*FeeResponse, error) {
	resp := &FeeResponse{}
	if err := self.queryWoAuth(pair, "fee", resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type TickerResponse struct {
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Avg        float64 `json:"avg"`
	Vol        float64 `json:"vol"`
	VolCur     float64 `json:"vol_cur"`
	Last       float64 `json:"last"`
	Buy        float64 `json:"buy"`
	Sell       float64 `json:"sell"`
	Updated    int64   `json:"updated"`
	ServerTime int64   `json:"server_time"`
}

//Get service fee, see example https://btc-e.com/api/2/btc_usd/ticker
func (self *Btce) GetTicker(pair Pair) (*TickerResponse, error) {
	respWrapp := make(map[string]json.RawMessage)
	if err := self.queryWoAuth(pair, "ticker", &respWrapp); err != nil {
		return nil, err
	}
	resp := &TickerResponse{}
	if err := json.Unmarshal(respWrapp["ticker"], &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type TradesResponse []Trade

type Trade struct {
	Created       int64   `json:"date"`
	Price         float64 `json:"price"`
	Amount        float64 `json:"amount"`
	Tid           int64   `json:"tid"`
	PriceCurrency string  `json:"price_currency"`
	Item          string  `json:"item"`
	TradeType     string  `json:"trade_type"`
}

//Get trades, see example https://btc-e.com/api/2/btc_usd/trades
func (self *Btce) GetTrades(pair Pair) (*TradesResponse, error) {
	resp := &TradesResponse{}
	if err := self.queryWoAuth(pair, "trades", resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type DepthResponse struct {
	Asks []Depth `json:"asks"`
	Bids []Depth `json:"bids"`
}

type Depth struct {
	Price  float64
	Amount float64
}

//Get depth, see example https://btc-e.com/api/2/btc_usd/depth
func (self *Btce) GetDepth(pair Pair) (*DepthResponse, error) {
	respWrapp := make(map[string]interface{})
	if err := self.queryWoAuth(pair, "depth", &respWrapp); err != nil {
		return nil, err
	}
	resp := &DepthResponse{}
	for key, items := range respWrapp {
		if key == "asks" {
			for _, values := range items.([]interface{}) {
				value := values.([]interface{})
				depth := Depth{value[0].(float64), value[1].(float64)}
				resp.Asks = append(resp.Asks, depth)
			}
		} else {
			for _, values := range items.([]interface{}) {
				value := values.([]interface{})
				depth := Depth{value[0].(float64), value[1].(float64)}
				resp.Bids = append(resp.Bids, depth)
			}
		}
	}

	return resp, nil
}
