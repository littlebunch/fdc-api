// Package ds provides an interface for application calls to the datastore.
// To add a data source simply implement the methods
package ds

import "github.com/littlebunch/gnutdata-bfpd-api/model"

// DataSource wraps the basic methods used for accessing and updating a
// data store.
type DataSource interface {
	ConnectDs(cs fdc.Config) error
	Get(q string, f interface{}) error
	Query(q string, f *[]interface{}) error
	Counts(bucket string, doctype string, c *[]interface{}) error
	GetDictionary(dsname string, doctype string, offset int64, limit int64, n *interface{}) error
	Browse(bucket string, where string, offset int64, limit int64, sort string, order string, f *[]interface{}) error
	Search(sr fdc.SearchRequest, foods *[]interface{}) (int, error)
	Update(id string, r interface{})
	Bulk(n *[]fdc.NutrientData) error
	CloseDs()
}
