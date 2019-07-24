// Package cb implements the DataStore interface for CouchBase
package cb

import (
	"encoding/json"
	"fmt"
	"log"

	fdc "github.com/littlebunch/gnutdata-bfpd-api/model"

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

// Performs an arbitrary but well-formed query
func (ds Cb) Query(q string, f *[]interface{}) error {
	query := gocb.NewN1qlQuery(q)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
	if err == nil {
		var row interface{}
		for rows.Next(&row) {
			*f = append(*f, row)
		}
	}
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

// Browse fills out a slice of Foods, Nutrients or NutrientData items, returns gocb error
func (ds *Cb) Browse(bucket string, where string, offset int64, limit int64, sort string, order string, f *[]interface{}) error {

	q := fmt.Sprintf("select * from %s  as food use index(%s) where %s is not missing and %s order by %s %s offset %d limit %d", bucket, useIndex(sort, order), sort, where, sort, order, offset, limit)
	fmt.Printf("QUERY=%s\n", q)
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
	switch sr.SearchType {
	case fdc.PHRASE:
		query = gocb.NewSearchQuery(sr.IndexName, cbft.NewMatchPhraseQuery(sr.Query).Field(sr.SearchField)).Limit(int(sr.Max)).Skip(sr.Page).Fields("*")
	case fdc.WILDCARD:
		query = gocb.NewSearchQuery(sr.IndexName, cbft.NewWildcardQuery(sr.Query).Field(sr.SearchField)).Limit(int(sr.Max)).Skip(sr.Page).Fields("*")
	default:
		query = gocb.NewSearchQuery(sr.IndexName, cbft.NewMatchQuery(sr.Query).Field(sr.SearchField)).Limit(int(sr.Max)).Skip(sr.Page).Fields("*")
	}
	result, err := ds.Conn.ExecuteSearchQuery(query)
	if err != nil {
		fmt.Println("Query Error:", err.Error())
		return 0, err
	}
	count = result.TotalHits()
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
			} else {
				*foods = append(*foods, fdc.SearchResult{Food: f})
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

// Bulk inserts a list of Nutrient Data items
func (ds *Cb) Bulk(items *[]fdc.NutrientData) error {
	var v []gocb.BulkOp
	for _, r := range *items {
		v = append(v, &gocb.InsertOp{Key: fmt.Sprintf("%s_%d", r.FdcID, r.Nutrientno), Value: r})
	}
	return ds.Conn.Do(v)

}

// Generates a use index phrase for use by Browse
// to speed up the sort
func useIndex(sort string, order string) string {
	useindex := ""
	switch sort {
	case "foodDescription":
		useindex = "idx_fd"
	case "company":
		useindex = "idx_company"
	case "fdcid":
	default:
		useindex = "idx_fdcId"
	}
	if order == "desc" {
		useindex = useindex + "_desc"
	} else {
		useindex = useindex + "_asc"
	}
	return useindex
}

/* Possible nutrient report:
*select  distinct meta(g).id,g.foodDescription,g.dataSource,n.nutrientNumber,n.valuePer100UnitServing from gnutdata n
join gnutdata g on n.fdcId = meta(g).id
where n.type="NUTDATA" and g.type="FOOD" and g.dataSource="SR"  and n.nutrientNumber=307 and n.valuePer100UnitServing is not missing
order by n.valuePer100UnitServing DESC
offset 0
Limit 100
*/
/* Possible only return nutrients no
 * select g.fdcId,g.foodDescription,
 * ARRAY n for n in g.nutrients when n.nutrientNumber=204 end as nutrients
 * from `gnutdata` as g
 * where  g.type="FOOD"
 * offset 0
 * limit 50;
 */
