package fdc

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

//Config provides basic CouchBase configuration properties for API services.  Properties are normally read in from a YAML file or the environment
type Config struct {
	CouchDb CouchDb
}

// CouchDb configuration for connecting, reading and writing Couchbase nodes
type CouchDb struct {
	URL    string
	Bucket string
	Fts    string
	User   string
	Pwd    string
}

// Defaults sets values for CouchBase configuration properties if none have been provided.
func (cs *Config) Defaults() {
	if os.Getenv("COUCHBASE_URL") != "" {
		cs.CouchDb.URL = os.Getenv("COUCHBASE_URL")
	}
	if os.Getenv("COUCHBASE_BUCKET") != "" {
		cs.CouchDb.Bucket = os.Getenv("COUCHBASE_BUCKET")
	}
	if os.Getenv("COUCHBASE_FTSINDEX") != "" {
		cs.CouchDb.Fts = os.Getenv("COUCHBASE_FTSINDEX")
	}
	if os.Getenv("COUCHBASE_USER") != "" {
		cs.CouchDb.User = os.Getenv("COUCHBASE_USER")
	}
	if os.Getenv("COUCHBASE_PWD") != "" {
		cs.CouchDb.Pwd = os.Getenv("COUCHBASE_PWD")
	}
	if cs.CouchDb.URL == "" {
		cs.CouchDb.URL = "localhost"
	}
	if cs.CouchDb.Bucket == "" {
		cs.CouchDb.Bucket = "gnutdata"
	}
	if cs.CouchDb.Fts == "" {
		cs.CouchDb.Fts = "fd_food"
	}
}

// GetConfig reads config from a file
func (cs *Config) GetConfig(c *string) {
	raw, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Println(err.Error())
	}
	if err = yaml.Unmarshal(raw, cs); err != nil {
		log.Println(err.Error())
	}
	cs.Defaults()
}
