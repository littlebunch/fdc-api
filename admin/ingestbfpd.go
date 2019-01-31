package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	fdc "github.com/littlebunch/gnutdata-bfpd-api/model"
)

type ingestCnt struct {
	Foods     int `json:"foods"`
	Servings  int `json:"servings"`
	Nutrients int `json:"nutrients"`
}

var (
	cnts ingestCnt
	err  error
	wg   sync.WaitGroup
)

// ProcessBFPDFiles loads a set of Branded Food Products csv files processed
// in this order:
//		Products.csv  -- main food file
//		Servings.csv  -- servings sizes for each food
//		Nutrients.csv -- nutrient values for each food
func ProcessBFPDFiles(path string) error {

	/*cnts.Foods, err = foods(path)
	if err != nil {
		log.Fatal(err)
	}*/
	//wg.Add(1)
	//go servings(path)
	wg.Add(1)
	go nutrients(path)
	wg.Wait()
	if err != nil {
		return err
	}
	log.Printf("Finished.  Counts: %d Foods %d Servings %d\n", cnts.Foods, cnts.Servings, cnts.Nutrients)
	return err
}
func foods(path string) (int, error) {
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

func servings(path string) (int, error) {
	defer wg.Done()
	fn := path + "branded_food.csv"
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

		id := *t + ":" + record[0]
		if cid != id {
			if cid != "" {
				food.Servings = s
				dc.Update(cid, food)
			}
			cid = id
			dc.Get(id, &food)
			food.Manufacturer = record[1]
			food.Upc = record[2]
			food.Ingredients = record[3]
			food.Source = record[8]
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
func nutrients(path string) (int, error) {
	defer wg.Done()
	fn := path + "food_nutrient.csv"
	f, err := os.Open(fn)
	if err != nil {
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

	nutmap := initNutrientInfoMap(il)
	for i := range nutmap {
		fmt.Printf("NAME=%v\n", nutmap[i].Name)
	}
	if err := dc.GetDictionary("gnutdata", "DERV", 0, 500, &il); err != nil {
		return 0, err
	}
	dlmap := initDerivationInfoMap(il)
	for i := range dlmap {
		fmt.Printf("%v\n", dlmap[i].Description)
	}
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

func initNutrientInfoMap(il interface{}) map[uint]fdc.Nutrient {
	m := make(map[uint]fdc.Nutrient)
	switch il := il.(type) {
	case []fdc.Nutrient:
		for _, value := range il {
			m[uint(value.NutrientID)] = value
		}
	}
	return m
}
