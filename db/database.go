// Implements Oracle database access.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/goracle.v2"

	_ "gopkg.in/goracle.v2"
)

type Globals struct {
	Release string
	Version int
	RAC 	bool
	CDB		bool
	Role    string
}

type Vdatabase struct {
	OpenMode string
	FlashbackOn string
	ForceLogging string
	ControlfileType string
	ProtectionMode string
	ProtectionLevel string
	SwitchoverStatus string
	DataGuardBroker string

}

type ShortDetails struct {
	Details Globals
	Database Vdatabase
	Instances []Instance
}

type Instance struct {
	InstanceNumber int
	InstanceName string
	HostName string
	Status string
	Parallel string
	ThreadNumber int
}

type Component struct {
	Name string
	Version string
	Status string
}

type Container struct {
	Name string
	ID int
	OpenMode string
}

type FullDetails struct {
	Details Globals
	Database Vdatabase
	Instances []Instance
	Registry []Component
	Containers []Container
}

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

func DisplayDetails(cn *Connect, short bool, full bool, pretty bool) (err error) {
	cp := NewConnectParams(cn)
	connString := cp.StringWithPassword()
	//fmt.Println("Connect string: " + connString)

	db, err := sql.Open("goracle", connString)
	if err != nil {
		return fmt.Errorf("database open error %v (%s)", err, connString)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("database ping error %v", err)
	}

	var details interface{}
	if full {
		details, err = printFullDetails(db)
	} else if short {
		details, err = printShortDetails(db)
	} else {
		details, err = printDetails(db)
	}
	if err != nil {
		return err
	}

	return writeJSON(details, pretty)
}

func printDetails(db *sql.DB) (*Globals, error) {
	var details Globals
	if err := getDetails(db, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

func printShortDetails(db *sql.DB) (*ShortDetails, error) {
	var details ShortDetails
	var err error

	if err = getDetails(db, &details.Details); err != nil {
		return nil, err
	}

	if err = queryDatabase(db, &details.Database); err != nil {
		return nil, err
	}

	if details.Instances, err = queryInstance(db); err != nil {
		return nil, err
	}

	return &details, nil
}

func printFullDetails(db *sql.DB) (*FullDetails, error) {
	var details FullDetails
	var err error

	if err = getDetails(db, &details.Details); err != nil {
		return nil, err
	}

	if err = queryDatabase(db, &details.Database); err != nil {
		return nil, err
	}

	if details.Instances, err = queryInstance(db); err != nil {
		return nil, err
	}

	if details.Registry, err = queryRegistry(db); err != nil {
		return nil, err
	}

	if details.Details.Version < 12 {
		details.Containers = []Container{}
	} else {
		if details.Containers, err = queryContainers(db); err != nil {
			return nil, err
		}
	}

	return &details, nil
}

func getDetails(db *sql.DB, d *Globals) (err error) {
	d.Release, d.Version, err = checkVersion(db)
	if err != nil {
		return err
	}

	d.RAC, err = checkRAC(db)
	if err != nil {
		return err
	}

	d.CDB, err = checkCDB(db, d.Version)
	if err != nil {
		return err
	}

	d.Role, err = getRole(db)
	if err != nil {
		return err
	}

	return nil
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

func queryDatabase(db *sql.DB, d *Vdatabase) error {
	stmt := `SELECT flashback_on, force_logging, controlfile_type, open_mode, 
				protection_mode, protection_level, switchover_status, dataguard_broker 
				FROM V$DATABASE`
	err := db.QueryRow(stmt).Scan(&d.FlashbackOn, &d.ForceLogging, &d.ControlfileType,
				&d.OpenMode, &d.ProtectionLevel, &d.ProtectionLevel, &d.SwitchoverStatus,
				&d.DataGuardBroker)
	if err != nil {
		return fmt.Errorf("query v$database: %v", err)
	}

	return nil
}

func queryInstance(db *sql.DB) ([]Instance, error) {
	stmt := `SELECT instance_number, instance_name, host_name, status, parallel,
				thread# from gv$instance order by instance_number`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("query v$instance: %v", err)
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		err := rows.Scan(&inst.InstanceNumber, &inst.InstanceName, &inst.HostName,
						&inst.Status, &inst.Parallel, &inst.ThreadNumber)
		if err != nil {
			return nil, fmt.Errorf("fetch instance rows: %v", err)
		}

		instances = append(instances, inst)
	}

	return instances, nil
}

func queryRegistry(db *sql.DB) ([]Component, error) {
	stmt := `SELECT comp_name, version, status from dba_registry order by comp_name`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("query dba_registry: %v", err)
	}
	defer rows.Close()

	var registry []Component
	for rows.Next() {
		var comp Component
		err := rows.Scan(&comp.Name, &comp.Version, &comp.Status)
		if err != nil {
			return nil, fmt.Errorf("fetch registry rows: %v", err)
		}

		registry = append(registry, comp)
	}

	return registry, nil
}

func queryContainers(db *sql.DB) ([]Container, error) {
	stmt := `SELECT name, con_id, open_mode from v$containers order by name`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("query v$containers: %v", err)
	}
	defer rows.Close()

	var containers []Container
	for rows.Next() {
		var cont Container
		err := rows.Scan(&cont.Name, &cont.ID, &cont.OpenMode)
		if err != nil {
			return nil, fmt.Errorf("fetch containers rows: %v", err)
		}

		containers = append(containers, cont)
	}

	return containers, nil
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

func writeJSON(db interface{}, pretty bool) (err error) {
	var data []byte

	if pretty {
		data, err = json.MarshalIndent(db, "", "\t")
	} else {
		data, err = json.Marshal(db)
	}
	if err != nil {
		return fmt.Errorf("writeJSON(): %v (db: %v)", err, db)
	}

	fmt.Printf("%s\n", data)

	return nil
}