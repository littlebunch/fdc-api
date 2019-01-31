// Package ds provides a wrapper for application calls to the datastore.
// A datastore, e.g. Couchbase, MongoDb, MariaDb, is implemented as a type assertions on
// the Conn element in the DS struct.  Thus, a datastore may be added by coding in the particulars
// as a case statement.
package ds

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/littlebunch/gnutdata-bfpd-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
	"gopkg.in/couchbase/gocb.v1/cbft"
)

// DS is the datastore connection interface
type DS struct {
	Conn interface{}
}

// ConnectDs connects to a datastore, e.g. Couchbase, MongoDb, etc.
func (ds *DS) ConnectDs(cs fdc.Config) error {
	var err error
	switch ds.Conn.(type) {
	case *gocb.Bucket:
		cluster, err := gocb.Connect("couchbase://" + cs.CouchDb.URL)
		if err != nil {
			log.Fatalln("Cannot connect to cluster ", err)
		}
		cluster.Authenticate(gocb.PasswordAuthenticator{
			Username: cs.CouchDb.User,
			Password: cs.CouchDb.Pwd,
		})
		ds.Conn, err = cluster.OpenBucket(cs.CouchDb.Bucket, "")
		if err != nil {
			log.Fatalln("Cannot connect to bucket!", err)
		}
	default:
		log.Fatalln("Invalid connection type!")
	}
	return err
}

// Get finds data for a single food
func (ds *DS) Get(q string, f interface{}) error {
	var err error
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		_, err = v.Get(q, &f)
	default:
		log.Fatalln("Invalid connection type!")
	}
	return err
}

func (ds *DS) Counts(bucket string, doctype string, c *[]interface{}) error {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		q := fmt.Sprintf("select type,count(*) as count from %s where type = '%s' group by type", bucket, doctype)

		query := gocb.NewN1qlQuery(q)
		rows, err := v.ExecuteN1qlQuery(query, nil)
		if err != nil {
			return err
		}
		var row interface{}
		for rows.Next(&row) {
			*c = append(*c, row)
		}
	default:
		log.Fatalln("Invalid connection")
	}
	return nil
}

// GetDictionary returns dictionary documents, e.g. food groups, nutrients, derivations, etc.
func (ds *DS) GetDictionary(bucket string, doctype string, offset int64, limit int64, n *interface{}) error {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		q := fmt.Sprintf("select gd.* from %s as gd where type='%s' offset %d limit %d", bucket, doctype, offset, limit)
		query := gocb.NewN1qlQuery(q)
		rows, err := v.ExecuteN1qlQuery(query, nil)
		if err != nil {
			return err
		}
		switch doctype {
		case "NUT":
			var i []fdc.Nutrient
			var row fdc.Nutrient
			for rows.Next(&row) {
				i = append(i, row)
			}
			*n = i
		case "DERV":
			var i []fdc.Derivation
			var row fdc.Derivation
			for rows.Next(&row) {
				i = append(i, row)
			}
			*n = i
		case "FGSR":
		case "FGFNDDS":
			var i []fdc.FoodGroup
			var row fdc.FoodGroup
			for rows.Next(&row) {
				i = append(i, row)
			}
			*n = i
		}
	default:
		log.Fatalln("Invalid connection type!")
	}
	return nil
}

// Browse fills out a slice of Foods items, returns error
func (ds *DS) Browse(bucket string, offset int64, limit int64, format string, sort string, f *[]interface{}) error {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		q := ""
		if format == fdc.FULL {
			q = fmt.Sprintf("select gd.* from %s as gd where %s != '' offset %d limit %d", bucket, sort, offset, limit)
		} else if format == fdc.NUTRIENTS {
			q = fmt.Sprintf("select fdcId,nutrients from %s where %s != '' offset %d limit %d", bucket, sort, offset, limit)
		} else if format == fdc.SERVING {
			q = fmt.Sprintf("select fdcId,servingSizes from %s where %s != '' offset %d limit %d", bucket, sort, offset, limit)
		} else {
			q = fmt.Sprintf("select gd.fdcId,gd.foodDescription,gd.upc,gd.company,gd.source,gd.ingredients from %s as gd where %s != '' offset %d limit %d", bucket, sort, offset, limit)
		}
		query := gocb.NewN1qlQuery(q)
		rows, err := v.ExecuteN1qlQuery(query, nil)
		if err != nil {
			return err
		}
		var row interface{}
		for rows.Next(&row) {
			*f = append(*f, row)
		}
	default:
		log.Fatalln("Invalid connection")
	}
	return nil
}

// Search performs a search query, fills out a Foods slice and returns count, error
func (ds *DS) Search(q string, f string, indexName string, format string, limit int, offset int, foods *[]interface{}) (int, error) {
	count := 0
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		var query *gocb.SearchQuery
		if f == "" {
			query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q)).Limit(int(limit)).Skip(offset).Fields("*")
		} else {
			query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q).Field(f)).Limit(int(limit)).Skip(offset).Fields("*")
		}
		result, err := v.ExecuteSearchQuery(query)
		count = result.TotalHits()
		if err != nil {
			return 0, err
		}
		f := fdc.FoodMeta{}
		for _, r := range result.Hits() {
			jrow, err := json.Marshal(&r.Fields)
			if err != nil {
				return 0, err
			}
			if err = json.Unmarshal(jrow, &f); err != nil {
				return 0, err
			}
			if format == fdc.META {
				*foods = append(*foods, f)
			} else {
				food := fdc.Food{}
				if _, err := v.Get(r.Id, &food); err != nil {
					return 0, err
				}
				if format == fdc.SERVING {
					*foods = append(*foods, fdc.SearchServings{Food: f, Servings: food.Servings})
				} else if format == fdc.NUTRIENTS {
					*foods = append(*foods, fdc.SearchNutrients{Food: f, Nutrients: food.Nutrients})
				} else {
					*foods = append(*foods, fdc.SearchResult{Food: f, Servings: food.Servings, Nutrients: food.Nutrients})
				}
			}
		}
	default:
		log.Fatalln("Invalid connection type!")
	}
	return count, nil
}

// Update updates an existing document in the datastore
func (ds *DS) Update(id string, r interface{}) {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		v.Upsert(id, r, 0)
	default:
		log.Fatalln("Invalid connection type!")
	}
}

// CloseDs is a wrapper for the connection close func
func (ds *DS) CloseDs() {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		v.Close()
	default:
		log.Fatalln("Invalid connection type!")
	}
}
