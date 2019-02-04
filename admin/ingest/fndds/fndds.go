package fndds

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/littlebunch/gnutdata-bfpd-api/admin/ingest"
	"github.com/littlebunch/gnutdata-bfpd-api/admin/ingest/dictionaries"
	"github.com/littlebunch/gnutdata-bfpd-api/ds"
	fdc "github.com/littlebunch/gnutdata-bfpd-api/model"
)

var (
	cnts ingest.Counts
	err  error
	wg   sync.WaitGroup
)

// ProcessFiles loads a set of FNDSS csv files processed
// in this order:
//		food.csv  -- main food file
//		food_portion.csv  -- servings sizes for each food
//		food_nutrient.csv -- nutrient values for each food
func ProcessFiles(path string, dc ds.DS, dt fdc.DocType) error {
	t := dt.ToString(fdc.BFPD)
	cnts.Foods, err = Foods(path, dc, &t)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(1)
	go Servings(path, dc, &t)
	wg.Add(1)
	go Nutrients(path, dc, &t)
	wg.Wait()
	if err != nil {
		return err
	}
	log.Printf("Finished.  Counts: %d Foods %d Servings %d\n", cnts.Foods, cnts.Servings, cnts.Nutrients)
	return err
}
func Foods(path string, dc ds.DS, t *string) (int, error) {
	cnt := 0
	fn := path + "food.csv"
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
	}
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
		dc.Update(*t+":"+record[0],
			fdc.Food{
				FdcID:           record[0],
				Description:     record[2],
				PublicationDate: pubdate,
				Type:            *t,
			})
	}
	return cnts.Foods, err
}

// Servings implements an ingest of fdc.Food.ServingSizes for FNDDS foods
func Servings(path string, dc ds.DS, t *string) (int, error) {
	defer wg.Done()
	fn := path + "food_portion.csv"
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
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
			return cnts.Servings, err
		}

		id := *t + ":" + record[1]
		if cid != id {
			if cid != "" {
				food.Servings = s
				dc.Update(cid, food)
			}
			cid = id
			dc.Get(id, &food)

			//food.Group = record[7]
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
	return cnts.Servings, err
}

// Nutrients implements an ingest of fdc.Food.NutrietData for FNDDS foods
func Nutrients(path string, dc ds.DS, t *string) (int, error) {
	defer wg.Done()
	fn := path + "food_nutrient.csv"
	f, err := os.Open(fn)
	if err != nil {
		log.Fatalf("Can't open input file %v", err)
		return 0, err
	}
	r := csv.NewReader(f)
	cid := ""
	var (
		food fdc.Food
		n    []fdc.NutrientData
		il   interface{}
	)
	if err := dc.GetDictionary("gnutdata", "NUT", 0, 500, &il); err != nil {
		return 0, err
	}
	nutmap := dictionaries.InitNutrientInfoMap(il)

	if err := dc.GetDictionary("gnutdata", "DERV", 0, 500, &il); err != nil {
		return 0, err
	}
	dlmap := dictionaries.InitDerivationInfoMap(il)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cnts.Nutrients, err
		}

		id := *t + ":" + record[1]
		if cid != id {
			if cid != "" {
				food.Nutrients = n
				dc.Update(cid, food)
			}
			cid = id
			dc.Get(id, &food)
			n = nil
		}
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
		n = append(n, fdc.NutrientData{
			Nutrientno: nutmap[uint(v)].Nutrientno,
			Value:      float32(w),
			Nutrient:   nutmap[uint(v)].Name,
			Unit:       nutmap[uint(v)].Unit,
			Derivation: fdc.Derivation{ID: dlmap[uint(d)].ID, Code: dlmap[uint(d)].Code, Description: dlmap[uint(d)].Description},
		})
		if cnts.Nutrients%30000 == 0 {
			log.Println("Nutrients Count = ", cnts.Nutrients)
		}

	}
	return cnts.Nutrients, err
}
func initDerivationInfoMap(il interface{}) map[uint]fdc.Derivation {
	m := make(map[uint]fdc.Derivation)
	switch il := il.(type) {
	case []fdc.Derivation:
		for _, value := range il {
			m[uint(value.ID)] = value
		}
	}
	return m
}
