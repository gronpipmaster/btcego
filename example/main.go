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
}
