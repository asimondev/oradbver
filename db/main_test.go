package db

import "testing"

func TestArgs(t *testing.T) {
	var tests = []struct {
		user string
		password string
		role string
		database string
		config string
		want bool
	} {
		{"", "", "", "", "", true},
		{"", "pwd", "", "", "", false},
		{"andrej", "andrej", "", "", "", true},
		{"sys", "oracle", "", "", "", true},
		{"sys", "oracle", "", "db1", "", true},
		{"", "oracle", "", "db1", "", false},
	}

	for _, test := range tests {
		cn := NewConnect(&test.user, &test.password, &test.role, &test.database, &test.config)
		if got := cn.CheckArgs() == nil; got != test.want {
			t.Errorf("Args: %#v  got = %v  wanted: %v", test, got, test.want)
		}
	}
}