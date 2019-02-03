package ingest

import (
	"github.com/littlebunch/gnutdata-bfpd-api/ds"
	"github.com/littlebunch/gnutdata-bfpd-api/model"
)

type IngestCnt struct {
	Foods     int `json:"foods"`
	Servings  int `json:"servings"`
	Nutrients int `json:"nutrients"`
}

type Ingest interface {
	ProcessFiles(path string, dc ds.DS, dt fdc.DocType) error
	Foods(path string, dc ds.DS, t *string) (int, error)
	Servings(path string, dc ds.DS, t *string) (int, error)
	Nutrients(path string, dc ds.DS, t *string) (int, error)
}
