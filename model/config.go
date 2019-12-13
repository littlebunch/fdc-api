package fdc

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

//Config provides basic configuration properties for API services.  Properties are normally read in from a YAML file or the environment
//Each datastore should have it's own type
type Config struct {
	CouchDb CouchDb
	Aws     Aws
}

// CouchDb configuration for connecting, reading and writing Couchbase nodes
type CouchDb struct {
	URL    string
	Bucket string
	Fts    string
	User   string
	Pwd    string
}

// Aws DynamoDB configuration for connecting, reading and writing Amazon DynamoDB tables
type Aws struct {
	Table  string // AWS DynamoDb table name
	Region string // AWS region
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
	if os.Getenv("AWS_DYNAMODB_TABLE") != "" {
		cs.Aws.Table = os.Getenv("AWS_DYNAMODB_TABLE")
	}
	if os.Getenv("AWS_DYNAMODB_REGION") != "" {
		cs.Aws.Table = os.Getenv("AWS_DYNAMODB_REGION")
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
