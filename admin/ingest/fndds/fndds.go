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

type Fndds struct {
	Doctype string
}

// ProcessFiles loads a set of FNDSS csv files processed
// in this order:
//		food.csv  -- main food file
//		food_portion.csv  -- servings sizes for each food
//		food_nutrient.csv -- nutrient values for each food
func (p Fndds) ProcessFiles(path string, dc ds.DS) error {
	cnts.Foods, err = foods(path, dc, p.Doctype)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(1)
	go servings(path, dc)
	wg.Add(1)
	go nutrients(path, dc)
	wg.Wait()
	if err != nil {
		return err
	}
	log.Printf("Finished.  Counts: %d Foods %d Servings %d Nutrients\n", cnts.Foods, cnts.Servings, cnts.Nutrients)
	return err
}
func foods(path string, dc ds.DS, t string) (int, error) {
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
		dc.Update(record[0],
			fdc.Food{
				FdcID:           record[0],
				Description:     record[2],
				PublicationDate: pubdate,
				Type:            t,
			})
	}
	return cnts.Foods, err
}

// Servings implements an ingest of fdc.Food.ServingSizes for FNDDS foods
func servings(path string, dc ds.DS) (int, error) {
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

		id := record[1]
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

		a, err := strconv.ParseFloat(record[7], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse serving amount " + record[3])
		}
		s = append(s, fdc.Serving{
			Nutrientbasis: "g",
			Description:   record[5],
			Servingamount: float32(a),
		})

	}
	return cnts.Servings, err
}

// Nutrients implements an ingest of fdc.Food.NutrietData for FNDDS foods
func nutrients(path string, dc ds.DS) (int, error) {
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

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cnts.Nutrients, err
		}

		id := record[1]
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
		var dv *fdc.Derivation
		dv = nil
		n = append(n, fdc.NutrientData{
			Nutrientno: nutmap[uint(v)].Nutrientno,
			Value:      float32(w),
			Nutrient:   nutmap[uint(v)].Name,
			Unit:       nutmap[uint(v)].Unit,
			Derivation: dv,
		})
		if cnts.Nutrients%30000 == 0 {
			log.Println("Nutrients Count = ", cnts.Nutrients)
		}

	}
	return cnts.Nutrients, err
}
