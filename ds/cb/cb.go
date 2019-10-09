// Package cb implements the DataStore interface for CouchBase
package cb

import (
	"encoding/json"
	"fmt"
	"log"

	fdc "github.com/littlebunch/fdc-api/model"

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
	q := fmt.Sprintf("SELECT dataSource,count(*) AS count from %s WHERE type='FOOD' AND dataSource = '%s' GROUP BY dataSource", bucket, doctype)
	return ds.Query(q, c)
}

// GetDictionary returns dictionary documents, e.g. food groups, nutrients, derivations, etc.
func (ds *Cb) GetDictionary(bucket string, doctype string, offset int64, limit int64) ([]interface{}, error) {
	var i []interface{}
	q := fmt.Sprintf("select gd.* from %s as gd where type='%s' offset %d limit %d", bucket, doctype, offset, limit)
	query := gocb.NewN1qlQuery(q)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return nil, err
	}
	switch doctype {
	case "NUT":
		var row fdc.Nutrient
		for rows.Next(&row) {
			i = append(i, row)
		}
	case "DERV":
		var row fdc.Derivation
		for rows.Next(&row) {
			i = append(i, row)
		}
	case "FGSR":
		var row fdc.FoodGroup
		for rows.Next(&row) {
			i = append(i, row)
		}
	case "FGFNDDS":
		var row fdc.FoodGroup
		for rows.Next(&row) {
			i = append(i, row)
		}
	}
	return i, nil
}

// Browse fills out a slice of Foods, Nutrients or NutrientData items, returns gocb error
func (ds *Cb) Browse(bucket string, where string, offset int64, limit int64, sort string, order string) ([]interface{}, error) {
	var (
		row fdc.Food
		f   []interface{}
	)
	q := fmt.Sprintf("select food.* from %s as food use index(%s) where %s is not missing and %s order by %s %s offset %d limit %d", bucket, useIndex(sort, order), sort, where, sort, order, offset, limit)
	query := gocb.NewN1qlQuery(q)
	query.Timeout(200000)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return f, err
	}
	for rows.Next(&row) {
		f = append(f, row)
	}
	return f, nil
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

// NutrientReport Runs a NutrientReportRequest
func (ds *Cb) NutrientReport(bucket string, nr fdc.NutrientReportRequest, nutrients *[]interface{}) error {
	w := ""
	if nr.Source == "BFPD" {
		w = fmt.Sprintf(" AND ( n.Datasource = '%s' OR n.Datasource='%s' )", "LI", "GDSN")
	} else if nr.Source != "" {
		w = fmt.Sprintf(" AND n.Datasource = '%s'", nr.Source)
	}
	n1ql := fmt.Sprintf("SELECT g.foodDescription,n.fdcId,n.nutrientNumber,n.valuePer100UnitServing,n.unit,n.Datasource FROM %s n USE index(idx_nutdata_query_desc) join gnutdata g on meta(g).id=n.fdcId WHERE n.type=\"NUTDATA\" AND n.nutrientNumber=%d AND n.valuePer100UnitServing between %f AND %f %s OFFSET %d LIMIT %d", bucket, nr.Nutrient, nr.ValueGTE, nr.ValueLTE, w, nr.Page, nr.Max)
	err := ds.Query(n1ql, nutrients)
	return err
}

// Update updates an existing document in the datastore using Upsert
func (ds *Cb) Update(id string, r interface{}) error {

	_, err := ds.Conn.Upsert(id, r, 0)
	return err

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
func (ds *Cb) BulkInsert(items []gocb.BulkOp) error {

	return ds.Conn.Do(items)

}

// Query performs an arbitrary but well-formed query
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
/* select n.fdcId,g.foodDescription,n.valuePer100UnitServing,n.unit from gnutdata n
join gnutdata g on meta(g).id = n.fdcId
where n.type="NUTDATA" and n.nutrientNumber=207
order by n.valuePer100UnitServing DESC
offset 0
limit 50
*/
