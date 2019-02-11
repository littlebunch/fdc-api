// loader parses food data central csv and ingests it into couchbase documents
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/littlebunch/gnutdata-bfpd-api/admin/ingest"
	"github.com/littlebunch/gnutdata-bfpd-api/admin/ingest/bfpd"
	"github.com/littlebunch/gnutdata-bfpd-api/admin/ingest/dictionaries"
	"github.com/littlebunch/gnutdata-bfpd-api/admin/ingest/fndds"
	"github.com/littlebunch/gnutdata-bfpd-api/ds"
	"github.com/littlebunch/gnutdata-bfpd-api/ds/cb"
	"github.com/littlebunch/gnutdata-bfpd-api/model"
)

var (
	c   = flag.String("c", "config.yml", "YAML Config file")
	l   = flag.String("l", "/tmp/ingest.out", "send log output to this file -- defaults to /tmp/ingest.out")
	i   = flag.String("i", "", "Input csv file")
	t   = flag.String("t", "", "Input file type")
	cnt = 0
	cs  fdc.Config
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
	var dt fdc.DocType
	dtype := dt.ToDocType(*t)
	if dtype == 999 {
		log.Fatalln("Valid t option is required")
	}

	var (
		cs fdc.Config
		in ingest.Ingest
		cb cb.Cb
		ds ds.DataSource
	)
	cs.GetConfig(c)
	// create a Couchbase datastore
	// and connect to it
	ds = &cb
	if err := ds.ConnectDs(cs); err != nil {
		log.Fatalln("Cannot connect to datastore ", err)
	}
	// implement the Ingest interface
	if dtype == fdc.BFPD {
		in = bfpd.Bfpd{Doctype: dt.ToString(fdc.BFPD)}
	} else if dtype == fdc.FNDDS {
		in = fndds.Fndds{Doctype: dt.ToString(fdc.FNDDS)}
	} else {
		in = dictionaries.Dictionary{Dt: dtype}
	}
	// ingest the csv files
	if err := in.ProcessFiles(*i, ds); err != nil {
		log.Fatal(err)
	}

	log.Println("Finished.")
	ds.CloseDs()
	os.Exit(0)
}
