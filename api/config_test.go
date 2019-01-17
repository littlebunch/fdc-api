package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/littlebunch/gnut-bfpd-api/model"
	"gopkg.in/yaml.v2"
)

var config = `
couchdb:
  url: localhost
  bucket: foods
  user: tester
  pwd: testpw
  fts: fd_food
`

// Configuration tests
func TestConfig(t *testing.T) {
	var cs fdc.Config

	if err := yaml.Unmarshal([]byte(config), &cs); err != nil {
		t.Errorf("Can't parse the config %v!", err)
	}
	rc, msg := chkConfig(&cs)
	if rc != true {
		t.Errorf(msg)
	}
}
func TestEnvConfig(t *testing.T) {

	os.Setenv("TEST_COUCHBASE_URL", "localhost")
	os.Setenv("TEST_COUCHBASE_BUCKET", "bfpd")
	os.Setenv("TEST_COUCHBASE_FTSINDEX", "foods")
	os.Setenv("TEST_COUCHBASE_USER", "tester")
	os.Setenv("TEST_COUCHBASE_PWD", "testpw")
	var cs fdc.Config
	cs.CouchDb.User = os.Getenv("TEST_COUCHBASE_USER")
	cs.CouchDb.Pwd = os.Getenv("TEST_COUCHBASE_PWD")
	cs.CouchDb.Bucket = os.Getenv("TEST_COUCHBASE_BUCKET")
	cs.CouchDb.FtsIndex = os.Getenv("TEST_COUCHBASE_FTSINDEX")
	rc, msg := chkConfig(&cs)
	if rc != true {
		t.Errorf(msg)
	}
	os.Setenv("TEST_COUCHBASE_URL", "")
	os.Setenv("TEST_COUCHBASE_BUCKET", "")
	os.Setenv("TEST_COUCHBASE_FTSINDEX", "")
	os.Setenv("TEST_COUCHBASE_USER", "")
	os.Setenv("TEST_COUCHBASE_PWD", "")
}

// test env override by invoking a fail by setting a bad password in the env
func TestEnvOverride(t *testing.T) {
	var cs fdc.Config
	var rc bool
	var msg string
	cc := []byte(config)
	yaml.Unmarshal(cc, &cs)
	chkConfig(&cs)
	os.Setenv("TEST_COUCHBASE_PWD", "testpwd")
	cs.CouchDb.Pwd = os.Getenv("TEST_COUCHBASE_PWD")
	rc, msg = chkConfig(&cs)
	if rc == true {
		t.Errorf("Should be false but got true %s", msg)
	} else {
		os.Setenv("TEST_COUCHBASE_PWD", "testpw")
		cs.CouchDb.Pwd = os.Getenv("TEST_COUCHBASE_PWD")
		rc, msg = chkConfig(&cs)
	}
	if rc == false {
		t.Errorf(msg)
	}
	os.Setenv("BFPD_PW_TEST", "")
}

// check to see if the config matches the values we've assigned
func chkConfig(cs *fdc.Config) (bool, string) {
	if "foods" != cs.CouchDb.Bucket {
		return false, "Wrong bucelet name " + cs.CouchDb.Bucket
	} else if "tester" != cs.CouchDb.User {
		return false, "Wrong user name " + cs.CouchDb.User
	} else if "testpw" != cs.CouchDb.Pwd {
		return false, "Wrong users password " + cs.CouchDb.Pwd
	}
	c := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8&parseTime=True&loc=Local", cs.CouchDb.User, cs.CouchDb.Pwd, cs.CouchDb.Bucket)
	if "tester:testpw@tcp(127.0.0.1:3306)/foods?charset=utf8&parseTime=True&loc=Local" != c {
		return false, "Connection string does not match"
	}
	return true, ""
}
