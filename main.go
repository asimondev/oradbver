package main

import (
	"flag"
	"fmt"

	//_ "gopkg.in/goracle.v2"

	"oradbver/db"
)

func main() {
	var u = flag.String("u", "", "database user")
	var p = flag.String("p", "", "user password")
	var r = flag.String("r", "", "system privilege")
	var d = flag.String("d", "", "database connect string")
	var cfg = flag.String("c", "", "JSON config file")
	var ping = flag.Bool("ping", false, "ping database")

	flag.Parse()

	cn := db.NewConnect(u, p, r, d, cfg)
	cn.CheckArgs()

	if err:= db.ConnectDatabase(cn); err != nil {
		fmt.Println(err)
	}

	if (*ping) {
		db.StartPinging(cn)
	}
}
