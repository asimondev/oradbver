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

	role, err := getRole(db)
	if err != nil {
		return err
	}

	return writeJSON(release, ver, rac, cdb, role)
}

func checkVersion(db *sql.DB) (string, int, error) {
	ver, err := goracle.ServerVersion(context.Background(), db)
	if err != nil {
		return "", 0, fmt.Errorf("goracle.ServerVerion() %v", err)
	}

	return fmt.Sprintf("%d.%d.%d.%d", ver.Version, ver.Release, ver.Update,
		ver.PortRelease), ver.Version, nil
}

func checkRAC(db *sql.DB) (bool, error) {
	stmt := `SELECT value FROM V$PARAMETER WHERE name = 'cluster_database'`
	var rac string

	err := db.QueryRow(stmt).Scan(&rac)
	if err != nil {
		return false, fmt.Errorf("checkRAC(): %v", err)
	}

	return rac == "TRUE", nil
}

func checkCDB(db *sql.DB, ver int) (bool, error) {
	if ver < 12 {
		return false, nil
	}

	var answer string
	stmt := `SELECT cdb FROM V$DATABASE`
	err := db.QueryRow(stmt).Scan(&answer)
	if err != nil {
		return false, fmt.Errorf("checkCDB(): %v", err)
	}

	return answer == "YES", nil
}

func getRole(db *sql.DB) (string, error) {
	var role string

	stmt := `select sys_context('userenv', 'database_role') from dual`
	err := db.QueryRow(stmt).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("getRole() %v", err)
	}

	return role, nil
}

func writeJSON(rel string, ver int, rac, cdb bool, role string) error {
	type OraVersion struct {
		Release string
		Version int
		RAC 	bool
		CDB		bool
		Role    string
	}

	var db = OraVersion{Release: rel, Version: ver, RAC: rac,
		CDB: cdb, Role: role}
	data, err := json.Marshal(db)
	if err != nil {
		return fmt.Errorf("writeJSON(): %v (db: %v)", err, db)
	}

	fmt.Printf("%s\n", data)

	return nil
}