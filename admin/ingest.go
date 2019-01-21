package main

// parses food data central csv into couchbase documents
import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/littlebunch/gnutdata-bfpd-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
)

var (
	c   = flag.String("c", "config.yml", "YAML Config file")
	l   = flag.String("l", "/tmp/ingest.out", "send log output to this file -- defaults to /tmp/ingest.out")
	m   = flag.Int("m", 20, "number of concurrent message handle -- defaults to 20")
	in  = flag.String("i", "", "Input csv file")
	t   = flag.String("t", "", "Input file type")
	cnt = 0
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
	// Connect to couchbase
	log.Println("URL=", cs.CouchDb.URL, " user=", cs.CouchDb.User, " pwd=", cs.CouchDb.Pwd)
	cluster, err := gocb.Connect("couchbase://" + cs.CouchDb.URL)
	if err != nil {
		log.Fatalln("Cannot connect to cluster ", err)
	}
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: cs.CouchDb.User,
		Password: cs.CouchDb.Pwd,
	})
	bucket, err := cluster.OpenBucket(cs.CouchDb.Bucket, "")
	if err != nil {
		log.Fatalln("Cannot connect to bucket!", err)
	}
	bucket.Manager("", "").CreatePrimaryIndex("", true, false)
	if dtype == fdc.BFPD {
		err := ProcessBFPDFiles(bucket, *in)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// read in the file and insert into couchbase
		f, err := os.Open(*in)
		if err != nil {
			log.Fatalf("Cannot open %s", *in)
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
				bucket.Upsert(*t+":"+record[1],
					fdc.FoodGroup{
						Code:        record[1],
						Type:        *t,
						Description: record[2],
						LastUpdate:  record[3],
					}, 0)
			case fdc.FGFNDDS:
				bucket.Upsert(*t+":"+record[0],
					fdc.FoodGroup{
						Code:        record[0],
						Type:        *t,
						Description: record[1],
					}, 0)
			case fdc.NUT:
				no, err := strconv.ParseInt(record[6], 10, 0)
				if err != nil {
					no = 0
				}
				bucket.Upsert(*t+":"+record[6],
					fdc.Nutrient{
						Nutrientno: uint(no),
						Tagname:    record[18],
						Name:       record[1],
						Unit:       record[2],
						Type:       *t,
					}, 0)

			}
		}
	}
	log.Println("Finished.")
	os.Exit(0)
}
