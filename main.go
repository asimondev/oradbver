package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/goracle.v2"
	_ "gopkg.in/goracle.v2"
	"log"
	"os"
	"strings"
)

func main() {
	var u = flag.String("u", "", "database user")
	var p = flag.String("p", "", "user password")
	var r = flag.String("r", "", "system privilege")

	flag.Parse()

	usr, pwd, role := checkArgs(*u, *p, *r)
	fmt.Printf("User: %s Password: %s Role: %s\n", usr, pwd, role)

	if len(role) > 0 {
		role = "sysdba=1&"
	}
	sessions := 1

	con := fmt.Sprintf("oracle://?%spoolMinSessions=%d& poolMaxSessions=%d& poolInrement=0",
		role, sessions, sessions)
	db, err := sql.Open("goracle", con)
	if err != nil {
		fmt.Printf("Error: database connect error %v (%s).", err, con)
		os.Exit(1)
	}
	defer db.Close()

	release, ver := checkVersion(db)
	writeJSON(release, ver, checkRAC(db), checkCDB(db, ver))
}

func checkArgs(usr, pwd, role string) (string, string, string) {
	role = strings.ToLower(role)

	if role != "sysdba" && role != "" {
		log.Fatal("Error: unknown system privilege " + role)
	}

	if len(usr) > 0 {
		if len(pwd) == 0 {
			log.Fatal("Error: password is missing.")
		}
	} else {
		if len(pwd) > 0 {
			log.Fatal("Error: password specified for unknown user.")
		}

		if len(role) == 0 {
			role = "sysdba"
		}
	}

	return usr, pwd, role
}

func checkVersion(db *sql.DB) (string, int) {
	ver, err := goracle.ServerVersion(db)
	if err != nil {
		fmt.Printf("Error: ServerVerion() %v\n", err)
		os.Exit(1)
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