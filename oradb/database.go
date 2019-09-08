package oradb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/goracle.v2"
	"log"
	"os"

	_ "gopkg.in/goracle.v2"
)

var ctx context.Context

func NewConnectParams(cn *Connect) *goracle.ConnectionParams {
	var cp goracle.ConnectionParams = goracle.ConnectionParams{
		StandaloneConnection: true,
		//MinSessions: 1,
		//MaxSessions: 1,
		//PoolIncrement: 0,
		//ConnClass: "POOLED",
	}

	if cn.role == "sysdba" {
		cp.IsSysDBA = true
	}

	if cn.user != "" {
		cp.Username = cn.user
	}

	if cn.password != "" {
		cp.Password = cn.password
	}

	if cn.database != "" {
		cp.SID = cn.database
	}

	return &cp
}

func ConnectDatabase(cn *Connect) {
	cp := NewConnectParams(cn)
	connString := cp.StringWithPassword()
	//fmt.Println("Connect string: " + connString)

	db, err := sql.Open("goracle", connString)
	if err != nil {
		log.Fatalf("Error: database open error %v (%s).", err, connString)
	}
	defer db.Close()

	if err:= db.Ping(); err != nil {
		log.Fatalf("Error: database ping error %v.", err)
	}

	release, ver := checkVersion(db)
	writeJSON(release, ver, checkRAC(db), checkCDB(db, ver))
}

func checkVersion(db *sql.DB) (string, int) {
	ver, err := goracle.ServerVersion(context.Background(), db)
	if err != nil {
		log.Fatalf("Error: ServerVerion() %v.\n", err)
	}

	return fmt.Sprintf("%d.%d.%d.%d", ver.Version, ver.Release, ver.Update,
		ver.PortRelease), ver.Version
}

func checkRAC(db *sql.DB) bool {
	stmt := `SELECT value FROM V$PARAMETER WHERE name = 'cluster_database'`

	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Printf("Error: checkRAC() %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var s string
	for rows.Next() {
		rows.Scan(&s)
	}

	if s != "TRUE" {
		return false
	}

	return true
}

func checkCDB(db *sql.DB, ver int) bool {
	if ver < 12 {
		return false
	}

	stmt := `SELECT cdb FROM V$DATABASE`
	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Printf("Error: checkCDB() %v\n", err)
		os.Exit(1)
	}

	defer rows.Close()

	var answer string
	for rows.Next() {
		rows.Scan(&answer)
	}

	if answer == "NO" {
		return false
	}

	return true
}

func writeJSON(rel string, ver int, rac, cdb bool) {
	type OraVersion struct {
		Release string
		Version int
		RAC 	bool
		CDB		bool
	}

	var db = OraVersion{Release: rel, Version: ver, RAC: rac, CDB: cdb}
	data, err := json.Marshal(db)
	if err != nil {
		fmt.Printf("Error: Marshal() %v (db: %v).\n", err, db)
		os.Exit(1)
	}
	fmt.Printf("%s\n", data)
}