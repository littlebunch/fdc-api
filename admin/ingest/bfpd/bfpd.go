// Package bfpd implements an Ingest for Branded Food Products
package bfpd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/littlebunch/fdc-api/admin/ingest"
	"github.com/littlebunch/fdc-api/admin/ingest/dictionaries"
	"github.com/littlebunch/fdc-api/ds"
	fdc "github.com/littlebunch/fdc-api/model"
)

var (
	cnts ingest.Counts
	err  error
)

// Bfpd for implementing the interface
type Bfpd struct {
	Doctype string
}

// ProcessFiles loads a set of Branded Food Products csv files processed
// in this order:
//		Products.csv  -- main food file
//		Servings.csv  -- servings sizes for each food
//		Nutrients.csv -- nutrient values for each food
func (p Bfpd) ProcessFiles(path string, dc ds.DataSource) error {
	var errs, errn error
	rcs, rcn := make(chan error), make(chan error)
	c1, c2 := true, true
	err := servings(path, dc)
	if err != nil {
		log.Fatal(err)
	}
	go foods(path, dc, p.Doctype, rcs)
	go nutrients(path, dc, rcn)
	for c1 || c2 {
		select {
		case errs, c1 = <-rcs:
			if c1 {
				if errs != nil {
					fmt.Printf("Error from foods: %v\n", errs)
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
		}
	}
	log.Printf("Finished.  Counts: %d Foods %d Servings %d Nutrients\n", cnts.Foods, cnts.Servings, cnts.Nutrients)
	return err
}
func foods(path string, dc ds.DataSource, t string, rc chan error) {
	defer close(rc)
	var food fdc.Food
	fn := path + "food.csv"
	f, err := os.Open(fn)
	if err != nil {
		rc <- err
		return
	}
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
		cnts.Foods++
		if cnts.Foods%1000 == 0 {
			log.Println("Count = ", cnts.Foods)
		}
		pubdate, err := time.Parse("2006-01-02", record[4])
		if err != nil {
			log.Println(err)
		}
		if err = dc.Get(record[0], &food); err != nil {
			fmt.Printf("Cannot fetch record %s", record[0])
		}
		food.Description = record[2]
		food.PublicationDate = pubdate

		dc.Update(record[0], food)
	}
	rc <- err
	return
}

func servings(path string, dc ds.DataSource) error {

	fn := path + "branded_food.csv"
	fgid := 0
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	r := csv.NewReader(f)
	cid := ""
	var (
		food fdc.Food
		s    []fdc.Serving
		dt   *fdc.DocType
	)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		id := record[0]
		if cid != id {
			if cid != "" {
				food.FdcID = cid
				food.Servings = s
				dc.Update(cid, food)
			}
			cid = id
			//dc.Get(id, &food)
			food.Manufacturer = record[1]
			food.Upc = record[2]
			food.Ingredients = record[3]
			food.Source = record[8]
			food.Type = dt.ToString(fdc.FOOD)
			if record[7] != "" {
				fgid++
				food.Group = &fdc.FoodGroup{ID: int32(fgid), Description: record[7], Type: "FGGPC"}
			} else {
				food.Group = nil
			}
			s = nil
		}

		cnts.Servings++
		if cnts.Servings%10000 == 0 {
			log.Println("Servings Count = ", cnts.Servings)
		}

		a, err := strconv.ParseFloat(record[4], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse serving amount " + record[3])
		}
		s = append(s, fdc.Serving{
			Nutrientbasis: record[5],
			Description:   record[6],
			Servingamount: float32(a),
		})

	}
	return err
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

		cnts.Nutrients++
		w, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse value " + record[4])
		}

		v, err := strconv.ParseInt(record[2], 0, 32)
		if err != nil {
			log.Println(record[0] + ": can't parse nutrient no " + record[1])
		}
		d, err := strconv.ParseInt(record[5], 0, 32)
		if err != nil {
			log.Println(record[5] + ": can't parse derivation no " + record[1])
		}
		var dv *fdc.Derivation
		if dlmap[uint(d)].Code != "" {
			dv = &fdc.Derivation{ID: dlmap[uint(d)].ID, Code: dlmap[uint(d)].Code, Type: dt.ToString(fdc.DERV), Description: dlmap[uint(d)].Description}
		} else {
			dv = nil
		}
		if cid != id {
			if err = dc.Get(id, &food); err != nil {
				log.Printf("Cannot find %s %v", id, err)
			}
			cid = id
			source = food.Source

		}

		n = append(n, fdc.NutrientData{
			FdcID:      id,
			Nutrientno: nutmap[uint(v)].Nutrientno,
			Value:      float32(w),
			Nutrient:   nutmap[uint(v)].Name,
			Unit:       nutmap[uint(v)].Unit,
			Derivation: dv,
			Type:       dt.ToString(fdc.NUTDATA),
			Source:     source,
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
