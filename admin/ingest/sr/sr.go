// Package sr implements an Ingest for Standard Release Legacy foods
package sr

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/littlebunch/gnutdata-api/admin/ingest"
	"github.com/littlebunch/gnutdata-api/admin/ingest/dictionaries"
	"github.com/littlebunch/gnutdata-api/ds"
	fdc "github.com/littlebunch/gnutdata-api/model"
)

var (
	cnts ingest.Counts
	err  error
)

// Sr for implementing the interface
type Sr struct {
	Doctype string
}

// ProcessFiles loads a set of Standard Release csv files
func (p Sr) ProcessFiles(path string, dc ds.DataSource) error {
	var errs, errn, errndb error
	rcndb, rcs, rcn := make(chan error), make(chan error), make(chan error)
	c1, c2, c3 := true, true, true
	cnts.Foods, err = foods(path, dc, p.Doctype)
	if err != nil {
		log.Fatal(err)
	}
	go ndbnoCw(path, dc, rcndb)
	go servings(path, dc, rcs)
	go nutrients(path, dc, rcn)
	for c1 || c2 || c3 {
		select {
		case errs, c1 = <-rcs:
			if c1 {
				if errs != nil {
					fmt.Printf("Error from servings: %v\n", errs)
				} else {
					fmt.Printf("Servings ingest complete.\n")
				}
			}

		case errn, c2 = <-rcn:
			if c2 {
				if err != nil {
					fmt.Printf("Error from nutrients: %v\n", errn)
				} else {
					fmt.Printf("Nutrient ingest complete.\n")
				}
			}
		case errndb, c3 = <-rcndb:
			if c3 {
				if errndb != nil {
					fmt.Printf("Error from ndbno crosswalk: %v\n", errndb)
				} else {
					fmt.Printf("ndbno crosswalk complete.\n")
				}
			}
		}
	}
	log.Printf("Finished.  Counts: %d Foods %d Servings %d Nutrients\n", cnts.Foods, cnts.Servings, cnts.Nutrients)
	return err
}
func foods(path string, dc ds.DataSource, t string) (int, error) {
	var il interface{}
	var dt *fdc.DocType
	dtype := dt.ToString(fdc.FGSR)
	fn := path + "food.csv"
	cnt := 0
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
	}
	err = dc.GetDictionary("gnutdata", dtype, 0, 500, &il)
	if err != nil {
		fmt.Printf("Cannot load food group dictionary")
		return 0, err
	}
	fgmap := dictionaries.InitFoodGroupInfoMap(il)
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cnt, err
		}
		cnts.Foods++
		if cnts.Foods%1000 == 0 {
			log.Println("Count = ", cnts.Foods)
		}
		pubdate, err := time.Parse("2006-01-02", record[4])
		if err != nil {
			log.Println(err)
		}
		f, _ := strconv.ParseInt(record[3], 0, 32)
		var fg *fdc.FoodGroup
		if fgmap[uint(f)].Code != "" {
			fg = &fdc.FoodGroup{ID: fgmap[uint(f)].ID, Code: fgmap[uint(f)].Code, Description: fgmap[uint(f)].Description, Type: dtype}

		} else {
			fg = nil
		}
		dc.Update(record[0],
			fdc.Food{
				FdcID:           record[0],
				Description:     record[2],
				PublicationDate: pubdate,
				Source:          t,
				Group:           fg,
				Type:            dt.ToString(fdc.FOOD),
			})
	}
	return cnts.Foods, err
}

func servings(path string, dc ds.DataSource, rc chan error) {
	defer close(rc)
	fn := path + "food_portion.csv"
	f, err := os.Open(fn)
	if err != nil {
		rc <- err
		return
	}
	r := csv.NewReader(f)
	cid := ""
	var (
		food fdc.Food
		s    []fdc.Serving
	)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rc <- err
			return
		}

		id := record[1]
		if cid != id {
			if cid != "" {
				food.Servings = s
				dc.Update(cid, food)
			}
			cid = id
			dc.Get(id, &food)
			s = nil
		}

		cnts.Servings++
		if cnts.Servings%10000 == 0 {
			log.Println("Servings Count = ", cnts.Servings)
		}

		a, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse serving amount " + record[3])
		}
		w, err := strconv.ParseFloat(record[7], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse serving weight " + record[7])
		}
		p, _ := strconv.ParseInt(record[8], 0, 32)

		s = append(s, fdc.Serving{
			Nutrientbasis: record[5],
			Description:   record[6],
			Servingamount: float32(a),
			Weight:        float32(w),
			Datapoints:    int32(p),
		})

	}
	rc <- err
	return
}
func ndbnoCw(path string, dc ds.DataSource, rc chan error) {
	defer close(rc)
	fn := path + "sr_legacy_food.csv"
	f, err := os.Open(fn)
	if err != nil {
		rc <- err
		return
	}

	var food fdc.Food
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rc <- err
			return
		}
		dc.Get(record[0], &food)
		food.NdbNo = record[1]
		dc.Update(record[0], food)

	}
	rc <- nil
	return
}
func nutrients(path string, dc ds.DataSource, rc chan error) {
	defer close(rc)
	var (
		dt          *fdc.DocType
		food        fdc.Food
		cid, source string
	)
	fn := path + "food_nutrient.csv"
	f, err := os.Open(fn)
	if err != nil {
		rc <- err
		return
	}
	r := csv.NewReader(f)
	var (
		n  []fdc.NutrientData
		il interface{}
	)
	if err := dc.GetDictionary("gnutdata", dt.ToString(fdc.NUT), 0, 500, &il); err != nil {
		rc <- err
		return
	}
	nutmap := dictionaries.InitNutrientInfoMap(il)

	if err := dc.GetDictionary("gnutdata", dt.ToString(fdc.DERV), 0, 500, &il); err != nil {
		rc <- err
		return
	}
	dlmap := dictionaries.InitDerivationInfoMap(il)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rc <- err
			return
		}

		id := record[1]
		if cid != id {
			if err = dc.Get(id, &food); err != nil {
				log.Printf("Cannot find %s %v", id, err)
			}
			cid = id
			source = food.Source

		}
		cnts.Nutrients++
		w, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse value " + record[4])
		}
		min, _ := strconv.ParseFloat(record[6], 32)

		max, _ := strconv.ParseFloat(record[7], 32)

		v, err := strconv.ParseInt(record[2], 0, 32)
		if err != nil {
			log.Println(record[0] + ": can't parse nutrient no " + record[1])
		}
		d, _ := strconv.ParseInt(record[5], 0, 32)

		p, _ := strconv.ParseInt(record[4], 0, 32)

		var dv *fdc.Derivation
		if dlmap[uint(d)].Code != "" {
			dv = &fdc.Derivation{ID: dlmap[uint(d)].ID, Code: dlmap[uint(d)].Code, Type: dt.ToString(fdc.DERV), Description: dlmap[uint(d)].Description}
		} else {
			dv = nil
		}

		n = append(n, fdc.NutrientData{
			FdcID:      id,
			Source:     source,
			Nutrientno: nutmap[uint(v)].Nutrientno,
			Value:      float32(w),
			Nutrient:   nutmap[uint(v)].Name,
			Unit:       nutmap[uint(v)].Unit,
			Derivation: dv,
			Datapoints: int(p),
			Min:        float32(min),
			Max:        float32(max),
			Type:       dt.ToString(fdc.NUTDATA),
		})
		if cnts.Nutrients%1000 == 0 {
			log.Println("Nutrients Count = ", cnts.Nutrients)
			err := dc.Bulk(&n)
			if err != nil {
				log.Printf("Bulk insert failed: %v\n", err)
			}
			n = nil
		}

	}
	rc <- nil
	return
}
