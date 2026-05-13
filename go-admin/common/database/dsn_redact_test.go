package database

import "testing"

func TestRedactDataSourceName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			"root:secret@tcp(127.0.0.1:3306)/db?charset=utf8",
			"root:***@tcp(127.0.0.1:3306)/db?charset=utf8",
		},
		{
			"user:p@ss:wrd@tcp(host:3306)/x",
			"user:***@tcp(host:3306)/x",
		},
		{
			"root@tcp(127.0.0.1:3306)/db",
			"root@tcp(127.0.0.1:3306)/db",
		},
		{
			"noprotocol",
			"noprotocol",
		},
	}
	for _, c := range cases {
		got := RedactDataSourceName(c.in)
		if got != c.want {
			t.Errorf("RedactDataSourceName(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}
