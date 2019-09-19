// Package ds provides an interface for application calls to the datastore.
// To add a data source simply implement the methods
package ds

import (
	fdc "github.com/littlebunch/fdc-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
)

// DataSource wraps the basic methods used for accessing and updating a
// data store.
type DataSource interface {
	ConnectDs(cs fdc.Config) error
	Get(q string, f interface{}) error
	Query(q string, f *[]interface{}) error
	Counts(bucket string, doctype string, c *[]interface{}) error
	GetDictionary(dsname string, doctype string, offset int64, limit int64) ([]interface{}, error)
	Browse(bucket string, where string, offset int64, limit int64, sort string, order string) ([]interface{}, error)
	Search(sr fdc.SearchRequest, foods *[]interface{}) (int, error)
	NutrientReport(bucket string, nr fdc.NutrientReportRequest, nutrients *[]interface{}) error
	Update(id string, r interface{}) error
	Bulk(n *[]fdc.NutrientData) error
	BulkInsert(v *[]gocb.BulkOp) error
	CloseDs()
}
