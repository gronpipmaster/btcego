Btcego
========================

Wrapper api for https://btc-e.com, Allows for the use of the Private and Public APIs from BTC-e.

### Docs ###
See http://godoc.org/github.com/gronpipmaster/btcego

### Example ###
```go
auth := btcego.Auth{
  AccessKey: "your-api-key-here",
  SecretKey: "your-api-secret-here",
}

btceInstance := btcego.New(auth)

testPair := btcego.Pair("btc_usd")

info, err := btceInstance.GetInfo()
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", info)

optionTrans := &btcego.TransHistoryRequest{Order: btcego.OrderDesc}
transHistory, err := btceInstance.TransHistory(optionTrans)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", transHistory)

optionTrade := &btcego.TradeHistoryRequest{Order: btcego.OrderDesc}
tradeHistory, err := btceInstance.TradeHistory(optionTrade)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", tradeHistory)

optionActiveOrders := &btcego.ActiveOrdersRequest{Pair: testPair}
activeOrders, err := btceInstance.ActiveOrders(optionActiveOrders)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", activeOrders)

fee, err := btceInstance.GetFee(testPair)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", fee)

ticker, err := btceInstance.GetTicker(testPair)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", ticker)

trade, err := btceInstance.GetTrades(testPair)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", trade)

depth, err := btceInstance.GetDepth(testPair)
if err != nil {
  log.Println(err)
}
log.Printf("%#v\n", depth)
```