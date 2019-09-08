package oradb

import (
	"fmt"
	"log"
	"strings"
)

type Connect struct {
	user string
	password string
	role string
	database string
}

func NewConnect(u, p, r, d *string) *Connect {
	return 	&Connect{
		user: *u, password: *p, role: *r, database: *d,
	}
}

func (c *Connect) CheckArgs() {
	c.role = strings.ToLower(c.role)

	if c.role != "" && c.role != "sysdba" && c.role != "sysbackup" {
		log.Fatalf("Error: unknown or unsupported system privilege '%s'.\n", *c)
	}

	if c.user != "" {
		if c.password == "" {
			log.Fatal("Error: password is missing.")
		}
	} else {
		if c.password != "" {
			log.Fatal("Error: password specified for unknown user.")
		}

		if c.role == "" {
			c.role = "sysdba"
		}
	}
}

func (c Connect) String() string {
	return fmt.Sprintf("Connect => user: '%s' password: '%s' role: '%s' database: '%s'.\n",
		c.user, c.password, c.role, c.database)
}


