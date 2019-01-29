package main

import (
	"encoding/csv"
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

	cnts.Foods, err = foods(path)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(1)
	go servings(path)
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
	fn := path + "Products.csv"
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
		update, err := time.Parse("2006-01-02 15:04:05", record[5])
		if err != nil {
			log.Println(err)
		}
		pubdate, err := time.Parse("2006-01-02 15:04:05", record[6])
		if err != nil {
			log.Println(err)
		}
		dc.Update(*t+":"+record[0],
			fdc.Food{
				FdcID:           record[0],
				Description:     record[1],
				Source:          record[2],
				Upc:             record[3],
				Manufacturer:    record[4],
				UpdatedAt:       update,
				PublicationDate: pubdate,
				Ingredients:     record[7],
				Type:            *t,
			})
	}
	return cnts.Foods, err
}
func servings(path string) (int, error) {
	defer wg.Done()
	fn := path + "Serving_size.csv"
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
			s = nil
		}

		cnts.Servings++
		if cnts.Servings%10000 == 0 {
			log.Println("Servings Count = ", cnts.Servings)
		}
		w, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse weight" + record[1])
		}
		a, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse serving amount " + record[3])
		}
		s = append(s, fdc.Serving{
			Nutrientbasis: record[2],
			Description:   record[4],
			Servingstate:  record[5],
			Weight:        float32(w),
			Servingamount: float32(a),
		})

	}
	return cnts.Servings, err
}
func nutrients(path string) (int, error) {
	defer wg.Done()
	fn := path + "Nutrients.csv"
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
	}
	r := csv.NewReader(f)
	cid := ""
	var (
		food fdc.Food
		n    []fdc.NutrientData
	)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cnts.Nutrients, err
		}

		id := *t + ":" + record[0]
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
		w, err := strconv.ParseFloat(record[4], 32)
		if err != nil {
			log.Println(record[0] + ": can't parse value " + record[4])
		}

		v, err := strconv.ParseInt(record[1], 0, 32)
		if err != nil {
			log.Println(record[0] + ": can't parse nutrient no " + record[1])
		}
		n = append(n, fdc.NutrientData{
			Nutrientno: uint(v),
			Nutrient:   record[2],
			Value:      float32(w),
			Derivation: record[3],
			Unit:       record[5],
		})
		if cnts.Nutrients%30000 == 0 {
			log.Println("Nutrients Count = ", cnts.Nutrients)
		}

	}
	return cnts.Nutrients, err
}
