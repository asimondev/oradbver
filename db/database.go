package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/goracle.v2"

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

	if cn.Role == "sysdba" {
		cp.IsSysDBA = true
	}

	if cn.User != "" {
		cp.Username = cn.User
	}

	if cn.Password != "" {
		cp.Password = cn.Password
	}

	if cn.Database != "" {
		cp.SID = cn.Database
	}

	return &cp
}

func ConnectDatabase(cn *Connect) error {
	cp := NewConnectParams(cn)
	connString := cp.StringWithPassword()
	//fmt.Println("Connect string: " + connString)

	db, err := sql.Open("goracle", connString)
	if err != nil {
		return fmt.Errorf("Error: Database open error %v (%s).", err, connString)
	}
	defer db.Close()

	if err:= db.Ping(); err != nil {
		return fmt.Errorf("Error: Database ping error %v.", err)
	}

	release, ver, err := checkVersion(db);
	if err != nil {
		return err
	}

	rac, err := checkRAC(db)
	if err != nil {
		return err
	}

	cdb, err := checkCDB(db, ver)
	if err != nil {
		return err
	}

	return writeJSON(release, ver, rac, cdb)
}

func checkVersion(db *sql.DB) (string, int, error) {
	ver, err := goracle.ServerVersion(context.Background(), db)
	if err != nil {
		return "", 0, fmt.Errorf("Error: ServerVerion() %v.", err)
	}

	return fmt.Sprintf("%d.%d.%d.%d", ver.Version, ver.Release, ver.Update,
		ver.PortRelease), ver.Version, nil
}

func checkRAC(db *sql.DB) (bool, error) {
	stmt := `SELECT value FROM V$PARAMETER WHERE name = 'cluster_database'`

	rows, err := db.Query(stmt)
	if err != nil {
		return false, fmt.Errorf("Error: checkRAC() %v", err)
	}
	defer rows.Close()

	var s string
	for rows.Next() {
		rows.Scan(&s)
	}

	if s != "TRUE" {
		return false, nil
	}

	return true, nil
}

func checkCDB(db *sql.DB, ver int) (bool, error) {
	if ver < 12 {
		return false, nil
	}

	stmt := `SELECT cdb FROM V$DATABASE`
	rows, err := db.Query(stmt)
	if err != nil {
		return false, fmt.Errorf("Error: checkCDB() %v", err)
	}

	defer rows.Close()

	var answer string
	for rows.Next() {
		rows.Scan(&answer)
	}

	if answer == "NO" {
		return false, nil
	}

	return true, nil
}

func writeJSON(rel string, ver int, rac, cdb bool) error {
	type OraVersion struct {
		Release string
		Version int
		RAC 	bool
		CDB		bool
	}

	var db = OraVersion{Release: rel, Version: ver, RAC: rac, CDB: cdb}
	data, err := json.Marshal(db)
	if err != nil {
		return fmt.Errorf("Error: Marshal() %v (db: %v).", err, db)
	}

	fmt.Printf("%s\n", data)

	return nil
}