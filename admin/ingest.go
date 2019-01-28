package main

// parses food data central csv into couchbase documents
import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/littlebunch/gnutdata-bfpd-api/ds"
	"github.com/littlebunch/gnutdata-bfpd-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
)

var (
	c   = flag.String("c", "config.yml", "YAML Config file")
	l   = flag.String("l", "/tmp/ingest.out", "send log output to this file -- defaults to /tmp/ingest.out")
	i   = flag.String("i", "", "Input csv file")
	t   = flag.String("t", "", "Input file type")
	cnt = 0
	b   *gocb.Bucket
	cs  fdc.Config
	dc  ds.DS
)

func init() {
	var (
		err   error
		lfile *os.File
	)

	lfile, err = os.OpenFile(*l, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", *l, ":", err)
	}
	m := io.MultiWriter(lfile, os.Stdout)
	log.SetOutput(m)
}
func main() {
	log.Print("Starting ingest")
	flag.Parse()
	dtype := fdc.ToDocType(*t)
	if dtype == 999 {
		log.Fatalln("Valid t option is required")
	}

	var cs fdc.Config
	cs.GetConfig(c)
	dc.Conn = b
	// connect to datastore
	if dc.ConnectDs(cs) != nil {
		log.Fatalln("Cannot connect to cluster ", err)
	}

	if dtype == fdc.BFPD {
		err := ProcessBFPDFiles(*i)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// read in the file and insert into couchbase
		f, err := os.Open(*i)
		if err != nil {
			log.Fatalf("Cannot open %s", *i)
		}
		r := csv.NewReader(f)
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			cnt++
			switch dtype {
			case fdc.FGSR:
				dc.Update(*t+":"+record[1],
					fdc.FoodGroup{
						Code:        record[1],
						Type:        *t,
						Description: record[2],
						LastUpdate:  record[3],
					})
			case fdc.FGFNDDS:
				dc.Update(*t+":"+record[0],
					fdc.FoodGroup{
						Code:        record[0],
						Type:        *t,
						Description: record[1],
					})
			case fdc.NUT:
				no, err := strconv.ParseInt(record[6], 10, 0)
				if err != nil {
					no = 0
				}
				dc.Update(*t+":"+record[6],
					fdc.Nutrient{
						Nutrientno: uint(no),
						Tagname:    record[18],
						Name:       record[1],
						Unit:       record[2],
						Type:       *t,
					})

			}
		}
	}
	log.Println("Finished.")
	os.Exit(0)
}
