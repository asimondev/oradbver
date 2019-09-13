package main

import (
	"flag"
	"fmt"
	"os"

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
	var pingOnce = flag.Bool("ping-once", false, "ping database once")

	flag.Parse()

	cn := db.NewConnect(u, p, r, d, cfg)
	cn.CheckArgs()

	if (*pingOnce) {
		os.Exit(db.PingOnce(cn))
	} else if (*ping) {
		db.StartPinging(cn)
	} else {
		err:= db.ConnectDatabase(cn)
		if err != nil {
			fmt.Println(err)
		}
	}
}
