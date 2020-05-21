// Package cb implements the DataStore interface for CouchBase
package cb

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

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

	case "FGFNDDS":
		fallthrough
	case "FGGPC":
		fallthrough
	case "FGSR":
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
		row interface{}
		f   []interface{}
	)
	q := fmt.Sprintf("select food.* from %s as food use index(%s) where %s is not missing and %s order by %s %s offset %d limit %d", bucket, useIndex(sort, order), sort, where, sort, order, offset, limit)
	query := gocb.NewN1qlQuery(q)
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
	var (
		sq     cbft.FtsQuery
		err    error
		result gocb.SearchResults
	)
	sr.Query = strings.Replace(sr.Query, "\"", "", -1)
	switch sr.SearchType {
	case fdc.PHRASE:
		sq = cbft.NewMatchPhraseQuery(sr.Query).Field(sr.SearchField)
	case fdc.WILDCARD:
		sq = cbft.NewWildcardQuery(sr.Query).Field(sr.SearchField)
	case fdc.REGEX:
		sq = cbft.NewRegexpQuery(sr.Query).Field(sr.SearchField)
	default:
		sq = cbft.NewMatchQuery(sr.Query).Field(sr.SearchField)
	}
	// add a foodgroup filter if we have one, otherwise run a standard search
	if sr.FoodGroup != "" {
		result, err = ds.Conn.ExecuteSearchQuery(gocb.NewSearchQuery(sr.IndexName, cbft.NewConjunctionQuery(sq, cbft.NewMatchQuery(sr.FoodGroup).Field("foodGroup.description"))).Limit(int(sr.Max)).Skip(sr.Page).Fields("*"))
	} else {
		result, err = ds.Conn.ExecuteSearchQuery(gocb.NewSearchQuery(sr.IndexName, sq).Limit(int(sr.Max)).Skip(sr.Page).Fields("*"))
	}
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
		*foods = append(*foods, f)
	}

	return count, nil
}

// NutrientReport Runs a NutrientReportRequest
func (ds *Cb) NutrientReport(bucket string, nr fdc.NutrientReportRequest, nutrients *[]interface{}) error {
	w := ""
	qfield := ""
	sort := "nutdata"

	if nr.FoodGroup != "" {
		w = fmt.Sprintf(" category=\"%s\" AND ", nr.FoodGroup)
		sort = "nutdata_fg"
	}
	if strings.ToLower(nr.Sort) == "portion" {
		sort = sort + "_portion"
		qfield = "n.portionValue"
	} else {
		qfield = "n.valuePer100UnitServing"
	}
	n1ql := fmt.Sprintf("SELECT n.foodDescription,n.upc,n.fdcId,n.category,n.company,n.valuePer100UnitServing,n.unit,n.portion,n.portionValue FROM %s n USE index(%s) WHERE %s n.type=\"NUTDATA\" AND n.nutrientNumber=%d AND %s between %f AND %f OFFSET %d LIMIT %d", bucket, useIndex(sort, nr.Order), w, nr.Nutrient, qfield, nr.ValueGTE, nr.ValueLTE, nr.Page, nr.Max)
	err := ds.Query(n1ql, nutrients)
	return err
}

// Update updates an existing document in the datastore using Upsert
func (ds *Cb) Update(id string, r interface{}) error {

	_, err := ds.Conn.Upsert(id, r, 0)
	return err

}

// Remove removes a document in the datastore
func (ds *Cb) Remove(id string) error {
	_, err := ds.Conn.Remove(id, 0)
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
		v = append(v, &gocb.InsertOp{Key: r.ID, Value: r})
	}
	return ds.Conn.Do(v)

}

// BulkInsert uses gocb library to insert a list of items defined in BulkOp struct
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

// FoodExists uses Couchbase subdoc API to determine if a key exists or not
func (ds Cb) FoodExists(id string) bool {
	rc := true
	_, err := ds.Conn.LookupIn(id).
		Exists("FdcID").Execute()
	if err != nil && err == gocb.ErrKeyNotFound {
		rc = false
	}
	return rc
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
	case "nutdata":
		useindex = "idx_nutdata_query"
	case "nutdata_portion":
		useindex = "idx_nutdata_portion_query"
	case "nutdata_fg_portion":
		useindex = "idx_nutdata_fg_portion_query"
	case "nutdata_fg":
		useindex = "idx_nutdata_fg_query"
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
