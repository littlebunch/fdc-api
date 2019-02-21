// Package cb implements the DataStore interface for CouchBase
package cb

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/littlebunch/gnutdata-bfpd-api/model"

	gocb "gopkg.in/couchbase/gocb.v1"
	"gopkg.in/couchbase/gocb.v1/cbft"
)

// Cb implements a DataSource interface to CouchBase
type Cb struct {
	Conn *gocb.Bucket
}

// ConnectDs connects to a datastore, e.g. Couchbase, MongoDb, etc.
func (ds *Cb) ConnectDs(cs fdc.Config) error {
	var err error
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
	return err
}

// Get finds data for a single food
func (ds Cb) Get(q string, f interface{}) error {
	_, err := ds.Conn.Get(q, &f)
	return err
}

// Counts returns document counts for a specified document type
func (ds *Cb) Counts(bucket string, doctype string, c *[]interface{}) error {
	q := fmt.Sprintf("select type,count(*) as count from %s where type = '%s' group by type", bucket, doctype)

	query := gocb.NewN1qlQuery(q)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return err
	}
	var row interface{}
	for rows.Next(&row) {
		*c = append(*c, row)
	}

	return nil
}

// GetDictionary returns dictionary documents, e.g. food groups, nutrients, derivations, etc.
func (ds *Cb) GetDictionary(bucket string, doctype string, offset int64, limit int64, n *interface{}) error {
	q := fmt.Sprintf("select gd.* from %s as gd where type='%s' offset %d limit %d", bucket, doctype, offset, limit)
	query := gocb.NewN1qlQuery(q)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
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
		var i []fdc.FoodGroup
		var row fdc.FoodGroup
		for rows.Next(&row) {
			i = append(i, row)
		}
		*n = i
	case "FGFNDDS":
		var i []fdc.FoodGroup
		var row fdc.FoodGroup
		for rows.Next(&row) {
			i = append(i, row)
		}
		*n = i
	}
	return nil
}

// Browse fills out a slice of Foods items, returns error
func (ds *Cb) Browse(bucket string, where string, offset int64, limit int64, format string, sort string, f *[]interface{}) error {

	q := ""
	if format == fdc.FULL {
		q = fmt.Sprintf("select gd.* from %s as gd where %s != ''  %s order by %s offset %d limit %d", bucket, sort, where, sort, offset, limit)
	} else if format == fdc.NUTRIENTS {
		q = fmt.Sprintf("select fdcId,nutrients from %s where %s != '' %s order by %s offset %d limit %d", bucket, sort, where, sort, offset, limit)
	} else if format == fdc.SERVING {
		q = fmt.Sprintf("select fdcId,servingSizes from %s where %s != '' %s orderby %s offset %d limit %d", bucket, sort, where, sort, offset, limit)
	} else {
		q = fmt.Sprintf("select gd.fdcId,gd.foodDescription,gd.upc,gd.company,gd.source,gd.ingredients from %s as gd where %s != '' %s order by %s offset %d limit %d", bucket, sort, where, sort, offset, limit)
	}
	query := gocb.NewN1qlQuery(q)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return err
	}
	var row interface{}
	for rows.Next(&row) {
		*f = append(*f, row)
	}

	return nil
}

// Search performs a search query, fills out a Foods slice and returns count, error
func (ds *Cb) Search(sr fdc.SearchRequest, foods *[]interface{}) (int, error) {
	count := 0

	var query *gocb.SearchQuery
	fmt.Printf("Query=%s", sr.Query)
	query = gocb.NewSearchQuery(sr.IndexName, cbft.NewMatchQuery(sr.Query)).Limit(int(sr.Max)).Skip(sr.Page).Fields("*")
	/*} else {
		query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q).Field(fld)).Limit(int(sr.Max)).Skip(sr.Offset).Fields("*")
	}*/
	result, err := ds.Conn.ExecuteSearchQuery(query)
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
		if sr.Format == fdc.META {
			*foods = append(*foods, f)
		} else {
			food := fdc.Food{}
			if _, err := ds.Conn.Get(r.Id, &food); err != nil {
				return 0, err
			}
			if sr.Format == fdc.SERVING {
				*foods = append(*foods, fdc.SearchServings{Food: f, Servings: food.Servings})
			} else if sr.Format == fdc.NUTRIENTS {
				*foods = append(*foods, fdc.SearchNutrients{Food: f, Nutrients: food.Nutrients})
			} else {
				*foods = append(*foods, fdc.SearchResult{Food: f, Servings: food.Servings, Nutrients: food.Nutrients})
			}
		}
	}

	return count, nil
}

// Update updates an existing document in the datastore using Upsert
func (ds *Cb) Update(id string, r interface{}) {

	ds.Conn.Upsert(id, r, 0)

}

// CloseDs is a wrapper for the connection close func
func (ds *Cb) CloseDs() {
	ds.Conn.Close()
}
