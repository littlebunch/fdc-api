package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	fdc "github.com/littlebunch/gnut-bfpd-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
)

type ingestCnt struct {
	Foods     int `json:"foods"`
	Servings  int `json:"servings"`
	Nutrients int `json:"nutrients"`
}

var (
	cnts ingestCnt
	err  error
)

// ProcessBFPDFiles loads a set of Branded Food Products csv files processed
// in this order:
//		Products.csv  -- main food file
//		Servings.csv  -- servings sizes for each food
//		Nutrients.csv -- nutrient values for each food
func ProcessBFPDFiles(bucket *gocb.Bucket, path string) (int, error) {
	//cnts.Foods, err = foods(bucket, path)
	cnts.Servings, err = servings(bucket, path)
	cnts.Nutrients, err = nutrients(bucket, path)
	if err != nil {
		return 0, err
	}
	return cnt, err
}
func foods(bucket *gocb.Bucket, path string) (int, error) {
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
		cnt++
		if cnt%1000 == 0 {
			log.Println("Count = ", cnt)
		}
		update, err := time.Parse("2006-01-02 15:04:05", record[5])
		if err != nil {
			log.Println(err)
		}
		pubdate, err := time.Parse("2006-01-02 15:04:05", record[6])
		if err != nil {
			log.Println(err)
		}
		bucket.Upsert(*t+":"+record[0],
			fdc.Food{
				FdcID:           record[0],
				Description:     record[1],
				Source:          record[2],
				Upc:             record[3],
				Manufacturer:    record[4],
				UpdatedAt:       update,
				PublicationDate: pubdate,
				Ingredients:     record[7],
			}, 0)
	}
	return cnt, err
}
func servings(bucket *gocb.Bucket, path string) (int, error) {
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
			return cnt, err
		}

		id := *t + ":" + record[0]
		if cid != id {
			if cid != "" {
				food.Servings = s
				bucket.Upsert(cid, food, 0)
			}
			cid = id
			s = nil
		}
		bucket.Get(id, &food)
		cnt++
		if cnt%10000 == 0 {
			log.Println("Servings Count = ", cnt)
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
	return cnt, err
}
func nutrients(bucket *gocb.Bucket, path string) (int, error) {
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
			return cnt, err
		}

		id := *t + ":" + record[0]
		if cid != id {
			if cid != "" {
				food.Nutrients = n
				bucket.Upsert(cid, food, 0)
			}
			cid = id
			n = nil
		}
		bucket.Get(id, &food)
		cnt++
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
		if cnt%30000 == 0 {
			log.Println("Nutrients Count = ", cnt)
		}

	}
	return cnt, err
}
