package dictionaries

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/littlebunch/gnutdata-bfpd-api/ds"
	"github.com/littlebunch/gnutdata-bfpd-api/model"
)

func ProcessFiles(path string, dc ds.DS, dtype fdc.DocType) error {
	// read in the file and insert into couchbase
	t := dtype.ToString(dtype)
	cnt := 0
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Cannot open %s", path)
		return err
	}
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("%v", err)
			return err
		}
		cnt++
		switch dtype {
		case fdc.DERV:
			id, err := strconv.ParseInt(record[0], 10, 0)
			if err != nil {
				id = 0
			}
			dc.Update(t+":"+record[0],
				fdc.Derivation{
					ID:          int32(id),
					Code:        record[1],
					Type:        t,
					Description: record[2],
				})
		case fdc.FGSR:
			no, err := strconv.ParseInt(record[0], 10, 0)
			if err != nil {
				no = 0
			}
			dc.Update(t+":"+record[0],
				fdc.FoodGroup{
					ID:          int32(no),
					Code:        record[1],
					Type:        t,
					Description: record[2],
					LastUpdate:  record[3],
				})
		case fdc.FGFNDDS:
			no, err := strconv.ParseInt(record[0], 10, 0)
			if err != nil {
				no = 0
			}
			dc.Update(t+":"+record[0],
				fdc.FoodGroup{
					ID:          int32(no),
					Type:        t,
					Description: record[1],
				})
		case fdc.NUT:
			var nid int64
			no, err := strconv.ParseInt(record[3], 10, 0)
			if err != nil {
				no = 0
			}
			nid, err = strconv.ParseInt(record[0], 10, 0)
			if err != nil {
				log.Println("Cannot parse ID: " + record[0])
				continue
			}
			dc.Update(t+":"+record[0],
				fdc.Nutrient{
					NutrientID: uint(nid),
					Nutrientno: uint(no),
					Name:       record[1],
					Unit:       record[2],
					Type:       t,
				})

		}
	}
	return nil
}
func InitNutrientInfoMap(il interface{}) map[uint]fdc.Nutrient {
	m := make(map[uint]fdc.Nutrient)
	switch il := il.(type) {
	case []fdc.Nutrient:
		for _, value := range il {
			m[uint(value.NutrientID)] = value
		}
	}
	return m
}
func InitDerivationInfoMap(il interface{}) map[uint]fdc.Derivation {
	m := make(map[uint]fdc.Derivation)
	switch il := il.(type) {
	case []fdc.Derivation:
		for _, value := range il {
			m[uint(value.ID)] = value
		}
	}
	return m
}
