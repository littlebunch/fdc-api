//Package ingest provides the Ingest interface
package ingest

import (
	"github.com/littlebunch/gnutdata-api/ds"
)

// Counts holds counts for documents loaded during an ingest
// process
type Counts struct {
	Foods     int `json:"foods"`
	Servings  int `json:"servings"`
	Nutrients int `json:"nutrients"`
	Other     int `json:"other"`
}

// Ingest wraps the basic methods used for loading different
// Food Data Central documents, i.e. Branded Foods, Standard release legacy,
// , Nutrients, etc..
type Ingest interface {
	ProcessFiles(path string, dc ds.DataSource) error
}
