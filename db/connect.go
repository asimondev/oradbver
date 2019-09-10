package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type Account struct {
	User     string
	Password string
	Role     string
	Database string
}

type Connect struct {
	Account
	config string
}

func NewConnect(u, p, r, d, cfg *string) *Connect {
	if *cfg != "" && (*u != "" || *p != "" || *d != "" || *r != "") {
		log.Fatal("Error: either config file or arguments are allowed.")
	}

	return 	&Connect{
		Account: Account{User: *u, Password: *p, Role: *r, Database: *d}, config: *cfg,
	}
}

func (c *Connect) readConfig(cfg string) {
	jsonFile, err := os.Open(c.config)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	data, _ := ioutil.ReadAll(jsonFile)
	if err = json.Unmarshal(data, &c.Account); err != nil {
		log.Fatal(err)
	}
}

func (c *Connect) CheckArgs() {
	if c.config != "" {
		c.readConfig(c.config)
	}

	c.Role = strings.ToLower(c.Role)

	if c.Role != "" && c.Role != "sysdba" && c.Role != "sysbackup" {
		log.Fatalf("Error: unknown or unsupported system privilege '%s'.\n", *c)
	}

	if c.User != "" {
		if c.Password == "" {
			fmt.Printf("Enter password: ")
			pwd, err := terminal.ReadPassword(0)
			if err != nil {
				log.Fatal("Error: Password is missing.")
			}
			fmt.Println()
			c.Password = string(pwd)
		}
	} else {
		if c.Password != "" {
			log.Fatal("Error: Password specified for unknown User.")
		}

		if c.Role == "" {
			c.Role = "sysdba"
		}
	}
}

func (c Connect) String() string {
	return fmt.Sprintf("Connect => User: '%s' Password: '%s' Role: '%s' Database: '%s'.\n",
		c.User, c.Password, c.Role, c.Database)
}


