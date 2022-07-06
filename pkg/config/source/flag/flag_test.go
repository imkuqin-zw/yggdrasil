package flag

import (
	"encoding/json"
	flag2 "flag"
	"testing"
)

var (
	dbuser = flag2.String("database-user", "default", "db user")
	dbhost = flag2.String("database-host", "", "db host")
	dbpw   = flag2.String("database-password", "", "db pw")
)

func initTestFlags() {
	flag2.Set("database-host", "localhost")
	flag2.Set("database-password", "some-password")
	flag2.Parse()
}

func TestFlagsrc_ReadAll(t *testing.T) {
	initTestFlags()
	source := NewSource()
	c, err := source.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data(), &actual); err != nil {
		t.Error(err)
	}
	actualDB := actual["database"].(map[string]interface{})

	// unset flag defaults should be loaded
	if actualDB["user"] != *dbuser {
		t.Errorf("expected %v got %v", *dbuser, actualDB["user"])
	}
}
