package btcego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const debug = true

const endpointUrl = "https://btc-e.com/tapi"

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
	// HTTP status code (200, 403, ...)
	StatusCode int
	Message    string `json:"error"`
}

func (self *Error) Error() string {
	return fmt.Sprintf("Error message:%s (StatusCode: %d)", self.Message, self.StatusCode)
}

type responseWrapper struct {
	Success int64           `json:"success"`
	Data    json.RawMessage `json:"return"`
}

func (self *Btce) query(params map[string]string, resp interface{}) error {
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

	if r.StatusCode != 200 {
		return buildError(r)
	}
	var respWrapper responseWrapper
	err = json.NewDecoder(r.Body).Decode(&respWrapper)
	if err != nil {
		return err
	}
	if respWrapper.Success == 0 {
		return buildError(r)
	}
	err = json.Unmarshal(respWrapper.Data, resp)
	if debug {
		fmt.Println("[--Response--]")
		fmt.Printf("%#v\n", resp)
	}
	return err
}

func multimap(p map[string]string) url.Values {
	q := make(url.Values, len(p))
	for k, v := range p {
		q[k] = []string{v}
	}
	return q
}

func buildError(r *http.Response) error {
	err := Error{}
	json.NewDecoder(r.Body).Decode(&err)
	err.StatusCode = r.StatusCode
	return &err
}

func makeParams(action string) map[string]string {
	params := make(map[string]string)
	params["method"] = action
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

type GetInfoResponse struct {
	Funds            Funds  `json:"funds"`
	Rights           Rights `json:"rights"`
	TransactionCount int64  `json:"transaction_count"`
	OpenOrders       int64  `json:"open_orders"`
	ServerTime       int64  `json:"server_time"`
}

//See https://btc-e.com/api/documentation section "getInfo"
func (self *Btce) GetInfo() (*GetInfoResponse, error) {
	params := makeParams("getInfo")
	self.nonce = self.nonce + 1
	resp := &GetInfoResponse{}
	if err := self.query(params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
