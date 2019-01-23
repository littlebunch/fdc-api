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

// Browse fills out a slice of Foods items
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

// CloseDs is a wrapper for the connection close func
func (ds *DS) CloseDs() {
	switch v := ds.Conn.(type) {
	case *gocb.Bucket:
		v.Close()
	default:
		log.Fatalln("Invalid connection type!")
	}
}
