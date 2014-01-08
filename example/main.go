package main

import (
	"github.com/gronpipmaster/btcego"
	"log"
)

func main() {
	auth := btcego.Auth{
		AccessKey: "your-api-key-here",
		SecretKey: "your-api-secret-here",
	}

	btceInstance := btcego.New(auth)

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

	optionActiveOrders := &btcego.ActiveOrdersRequest{Pair: "btc_usd"}
	activeOrders, err := btceInstance.ActiveOrders(optionActiveOrders)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v\n", activeOrders)

}
