package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

type DbDetails struct {
	Role string
	UniqueName string
	InstanceName string
	Server string
	Service string
}

func (d DbDetails) String() string {
	return fmt.Sprintf("Inst: %s  Host: %s  Service: %s  Db.Name: %s  Db.Role: %s",
		d.InstanceName, d.Server, d.Service, d.UniqueName, d.Role)
}

func PingDatabase(cn *Connect) {
	cp := NewConnectParams(cn)
	connString := cp.StringWithPassword()

	db, err := sql.Open("goracle", connString)
	if err != nil {
		log.Fatalf("Error: Database open error %v (%s).", err, connString)
	}
	defer db.Close()

	getSessionDetails(db)
}

func getSessionDetails(db *sql.DB) {
	stmt := `select sys_context('userenv', 'database_role'),
		sys_context('userenv', 'db_unique_name'),
		sys_context('userenv', 'instance_name'),
		sys_context('userenv', 'server_host'),
		sys_context('userenv', 'service_name')
		from dual`

	rows, err := db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var d DbDetails
	for rows.Next() {
		rows.Scan(&d.Role, &d.UniqueName, &d.InstanceName, &d.Server, &d.Service)
	}

	t := time.Now()
	fmt.Printf("%s  %s\n", t.Format("15:04:05"), d)
}

func StartPinging(cn *Connect) {
	abort := make(chan struct{})
	fmt.Println("\nPress Return to stop the pings...")
	go func(){
		os.Stdin.Read(make([]byte, 1)) // read one byte from input
		abort <- struct{}{}
	}()

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
			case <-ticker.C:
				PingDatabase(cn)
		case <-abort:
			ticker.Stop()
			return
		}
	}
}