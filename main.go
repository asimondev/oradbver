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

type Database struct {
	User string
	Password string
	Role string
	Connect string
}

func NewDatabase(u, p, r, c *string) Database {
	return 	Database{
		User: *u, Password: *p, Role: *r, Connect: *c,
	}
}

func (d *Database) ReadArgs() {
	d.Role = strings.ToLower(d.Role)

	if d.Role != "sysdba" && d.Role != "sysbackup" && d.Role != "" {
		log.Fatal("Error: unknown system privilege " + d.Role)
	}

	if d.User != "" {
		if d.Password == "" {
			log.Fatal("Error: password is missing.")
		}
	} else {
		if d.Password != "" {
			log.Fatal("Error: password specified for unknown user.")
		}

		if d.Role == "" {
			d.Role = "sysdba"
		}
	}
}

func (d Database) String() string {
	return fmt.Sprintf("Database => User: %s Password: %s Role: %s Connect: %s\n",
		d.User, d.Password, d.Role, d.Connect)
}

func (d *Database) ConnectString() string {
	var login string
	if d.User != "" {
		login = d.User + ":" + d.Password
	}

	var role string
	if d.Role == "sysdba" || d.User == "" && d.Role == "sysdba" {
		role = "sysdba=1& "
	}
	sessions := 1

	return fmt.Sprintf("oracle://%s?%spoolMinSessions=%d& " +
		"poolMaxSessions=%d& poolInrement=0", login, role, sessions, sessions)
}

func main() {
	var u = flag.String("u", "", "database user")
	var p = flag.String("p", "", "user password")
	var r = flag.String("r", "", "system privilege")
	var c = flag.String("d", "", "connect string")

	flag.Parse()

	d := NewDatabase(u, p, r, c)
	d.ReadArgs()
	fmt.Println(d)

	con := d.ConnectString()
	fmt.Println(con)
	db, err := sql.Open("goracle", con)
	if err != nil {
		fmt.Printf("Error: database connect error %v (%s).", err, con)
		os.Exit(1)
	}
	defer db.Close()

	release, ver := checkVersion(db)
	writeJSON(release, ver, checkRAC(db), checkCDB(db, ver))
}

//func checkArgs(usr, pwd, role string) (string, string, string) {
//	role = strings.ToLower(role)
//
//	if role != "sysdba" && role != "" {
//		log.Fatal("Error: unknown system privilege " + role)
//	}
//
//	if usr != "" {
//		if pwd == "" {
//			log.Fatal("Error: password is missing.")
//		}
//	} else {
//		if pwd != "" {
//			log.Fatal("Error: password specified for unknown user.")
//		}
//
//		if role == "" {
//			role = "sysdba"
//		}
//	}
//
//	return usr, pwd, role
//}

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