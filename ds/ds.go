// Package ds provides a wrapper for application calls to the datastore.
// A datastore, e.g. Couchbase, MongoDb, MariaDb, is implemented as a type assertions on
// the Conn element in the DS struct.  Thus, a datastore may be added by coding in the particulars
// as a case statement.
package ds

import (
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

		fmt.Printf("URL=%s\n", cs.CouchDb.URL)
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

// Browse fills out a slice of Foods items, returns error
func (ds *DS) Browse(bucket string, offset int64, limit int64, format string, sort string, f *[]interface{}) error {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		q := fmt.Sprintf("select * from %s as gd where %s != '' offset %d limit %d", bucket, sort, offset, limit)
		query := gocb.NewN1qlQuery(q)
		rows, err := v.ExecuteN1qlQuery(query, nil)
		if err != nil {
			return err
		}
		if format == fdc.META {
			type n1q struct {
				Item fdc.FoodMeta `json:"gd"`
			}
			var row n1q
			for rows.Next(&row) {
				*f = append(*f, row.Item)
			}
		} else {
			type n1q struct {
				Item fdc.Food `json:"gd"`
			}
			var row n1q
			for rows.Next(&row) {
				if format == fdc.SERVING {
					*f = append(*f, fdc.BrowseServings{FdcID: row.Item.FdcID, Servings: row.Item.Servings})
				} else if format == fdc.NUTRIENTS {
					*f = append(*f, fdc.BrowseNutrients{FdcID: row.Item.FdcID, Nutrients: row.Item.Nutrients})
				} else {
					*f = append(*f, row.Item)
				}
				row = n1q{}
			}
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
			query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q)).Limit(int(limit)).Skip(offset)
		} else {
			query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q).Field(f)).Limit(int(limit)).Skip(offset)
		}
		result, err := v.ExecuteSearchQuery(query)
		if err != nil {
			return 0, err
		}
		count = result.TotalHits()
		if format == fdc.META {
			var f fdc.FoodMeta
			for _, r := range result.Hits() {
				_, err := v.Get(r.Id, &f)
				if err != nil {
					return 0, err
				} else {
					*foods = append(*foods, f)
				}

			}
		} else {
			var f fdc.Food
			for _, r := range result.Hits() {
				_, err := v.Get(r.Id, &f)
				if err != nil {
					return 0, err
				} else {
					if format == fdc.SERVING {
						*foods = append(*foods, fdc.BrowseServings{FdcID: f.FdcID, Servings: f.Servings})
					} else if format == fdc.NUTRIENTS {
						*foods = append(*foods, fdc.BrowseNutrients{FdcID: f.FdcID, Nutrients: f.Nutrients})
					} else {
						*foods = append(*foods, f)
					}
					f = fdc.Food{}
				}

			}
		}
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
