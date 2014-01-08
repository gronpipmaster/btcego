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
		log.Fatalln(err)
	}
	log.Printf("%#v\n", info)

	optionTrans := &btcego.TransHistoryRequest{Order: btcego.OrderDesc}
	transHistory, err := btceInstance.TransHistory(optionTrans)
	if err != nil {
		log.Fatalln("transHistory err:", err)
	}
	log.Printf("%#v\n", transHistory)

	optionTrade := &btcego.TradeHistoryRequest{Order: btcego.OrderDesc}
	tradeHistory, err := btceInstance.TradeHistory(optionTrade)
	if err != nil {
		log.Fatalln("tradeHistory err:", err)
	}
	log.Printf("%#v\n", tradeHistory)
}
