package main

import (
	"flag"
	"log"
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
	var short = flag.Bool("short", false, "short database details")
	var full = flag.Bool("full", false, "full database details")
	var pretty = flag.Bool("pretty", false, "JSON pretty format output")

	flag.Parse()

	cn := db.NewConnect(u, p, r, d, cfg)
	if err := cn.CheckArgs(); err != nil {
		log.Fatal("Error: ", err)
	}

	if (*pingOnce) {
		os.Exit(db.PingOnce(cn))
	} else if (*ping) {
		db.StartPinging(cn)
	} else {
		err:= db.DisplayDetails(cn, *short, *full, *pretty)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}
}
