package db

import (
	"database/sql"
	"fmt"
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

	t := time.Now()
	fmt.Printf("%s  ", t.Format("15:04:05"))

	db, err := sql.Open("goracle", connString)
	if err != nil {
		fmt.Printf("database open error %v (%s).", err, connString)
		return
	}
	defer db.Close()

	getSessionDetails(db)
}


func getSessionDetails(db *sql.DB) {
	stmt := `select sys_context('userenv', 'db_unique_name'),
		sys_context('userenv', 'instance_name'),
		sys_context('userenv', 'server_host'),
		sys_context('userenv', 'service_name')
		from dual`

	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Printf("database query error %v", err)
		return
	}
	defer rows.Close()

	var d DbDetails
	for rows.Next() {
		rows.Scan(&d.UniqueName, &d.InstanceName, &d.Server, &d.Service)
	}

	d.Role, err = getRole(db)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%s", d)
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
				fmt.Println()
		case <-abort:
			ticker.Stop()
			return
		}
	}
}