// Package cdb implements the DataStore interface for Apache CouchDB
package cdb

import (
	"context"
	"fmt"
	"log"

	kivik "github.com/flimzy/kivik"
	_ "github.com/go-kivik/couchdb"
	fdc "github.com/littlebunch/fdc-api/model"
	"gopkg.in/couchbase/gocb.v1"
)

// Cdb implements a DataSource interface to CouchBase
type Cdb struct {
	Conn *kivik.DB
}

// ConnectDs connects to a datastore, e.g. Couchbase, MongoDb, etc.
func (ds *Cdb) ConnectDs(cs fdc.Config) error {
	var err error
	url := fmt.Sprintf("http://%s:%s@%s", cs.CouchDb.User, cs.CouchDb.Pwd, cs.CouchDb.URL)
	log.Println(url)
	conn, err := kivik.New(context.TODO(), "couch", url)
	if err != nil {
		log.Fatalln("Cannot get a client ", err)
	}
	ds.Conn, err = conn.DB(context.TODO(), cs.CouchDb.Bucket)
	if err != nil {
		log.Fatalln("Cannot connect to datastore!", err)
	}
	return err
}

// Get finds data for a single food
func (ds Cdb) Get(q string, f interface{}) error {
	r, err := ds.Conn.Get(context.TODO(), q)
	if err != nil {
		log.Fatalln("Get failed ", err)
	}
	return r.ScanDoc(&f)
}

// Counts returns document counts for a specified document type
func (ds *Cdb) Counts(bucket string, doctype string, c *[]interface{}) error {
	q := fmt.Sprintf("SELECT dataSource,count(*) AS count from %s WHERE type='FOOD' AND dataSource = '%s' GROUP BY dataSource", bucket, doctype)
	return ds.Query(q, c)
}

// GetDictionary returns dictionary documents, e.g. food groups, nutrients, derivations, etc.
func (ds *Cdb) GetDictionary(bucket string, doctype string, offset int64, limit int64) ([]interface{}, error) {

	var i []interface{}
	q := fmt.Sprintf("{\"selector\":{\"type\":\"%s\"},\"fields\":[],\"limit\":%d,\"skip\":%d}", doctype, limit, offset)
	rows, err := ds.Conn.Find(context.Background(), q)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}
	switch doctype {
	case "NUT":
		var row fdc.Nutrient
		for rows.Next() {
			rows.ScanDoc(&row)
			i = append(i, row)
		}
	case "DERV":
		var row fdc.Derivation
		for rows.Next() {
			rows.ScanDoc(&row)
			i = append(i, row)
		}
	case "FGSR":
		var row fdc.FoodGroup
		for rows.Next() {
			rows.ScanDoc(&row)
			i = append(i, row)
		}
	case "FGFNDDS":
		var row fdc.FoodGroup
		for rows.Next() {
			rows.ScanDoc(&row)
			i = append(i, row)
		}
	}

	//curl -H "Content-type:application/json" -X POST http://localhost:5984/srlegacy/_find -d '{"selector":{"type":{"$eq":"DERV"}},"fields":["id","code","description"],"limit":200,"skip":0}'
	return i, nil
}

// Browse fills out a slice of Foods, Nutrients or NutrientData items, returns gocb error
func (ds *Cdb) Browse(bucket string, where string, offset int64, limit int64, sort string, order string) ([]interface{}, error) {
	/*var (
		row fdc.Food
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
	}*/
	return nil, nil
}

// Search performs a search query, fills out a Foods slice and returns count, error
func (ds *Cdb) Search(sr fdc.SearchRequest, foods *[]interface{}) (int, error) {
	count := 0
	/*selector, err := mango.New(sr)
	if err != nil {
		return 0, err
	}*/
	/*var query *gocb.SearchQuery
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
	}*/

	return count, nil
}

// NutrientReport Runs a NutrientReportRequest
func (ds *Cdb) NutrientReport(bucket string, nr fdc.NutrientReportRequest, nutrients *[]interface{}) error {
	/*w := ""
	if nr.Source == "BFPD" {
		w = fmt.Sprintf(" AND ( n.Datasource = '%s' OR n.Datasource='%s' )", "LI", "GDSN")
	} else if nr.Source != "" {
		w = fmt.Sprintf(" AND n.Datasource = '%s'", nr.Source)
	}*/
	//n1ql := fmt.Sprintf("SELECT g.foodDescription,n.fdcId,n.nutrientNumber,n.valuePer100UnitServing,n.unit,n.Datasource FROM %s n USE index(idx_nutdata_query_desc) join gnutdata g on meta(g).id=n.fdcId WHERE n.type=\"NUTDATA\" AND n.nutrientNumber=%d AND n.valuePer100UnitServing between %f AND %f %s OFFSET %d LIMIT %d", bucket, nr.Nutrient, nr.ValueGTE, nr.ValueLTE, w, nr.Page, nr.Max)
	//err := ds.Query(n1ql, nutrients)
	return nil
}

// Update updates an existing document in the datastore using Upsert
func (ds *Cdb) Update(id string, r interface{}) error {
	rev, err := ds.Conn.Put(context.TODO(), id, r)
	if err != nil {
		log.Fatalln("Error on update: ", err)
	}
	log.Printf("Update %s rev %s\n", id, rev)
	//	_, err := ds.Conn.Upsert(id, r, 0)
	return nil

}

// CloseDs is a wrapper for the connection close func
func (ds *Cdb) CloseDs() {
	//ds.Conn.Close()
}

// Bulk inserts a list of Nutrient Data items
func (ds *Cdb) Bulk(items *[]fdc.NutrientData) error {
	rev, err := ds.Conn.BulkDocs(context.TODO(), items)
	if err != nil {
		log.Fatalln("Bulk insert error", err)
	}
	log.Printf("%v", rev)
	/*var v []gocb.BulkOp
	for _, r := range *items {
		v = append(v, &gocb.InsertOp{Key: fmt.Sprintf("%s_%d", r.FdcID, r.Nutrientno), Value: r})
	}
	return ds.Conn.Do(v)
	*/
	return nil

}
func (ds *Cdb) BulkInsert(items []gocb.BulkOp) error {

	//return ds.Conn.Do(items)
	return nil

}

// Query performs an arbitrary but well-formed query
func (ds Cdb) Query(q string, f *[]interface{}) error {
	/*query := gocb.NewN1qlQuery(q)
	rows, err := ds.Conn.ExecuteN1qlQuery(query, nil)
	if err == nil {
		var row interface{}
		for rows.Next(&row) {
			*f = append(*f, row)
		}
	}
	return err*/
	return nil
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
