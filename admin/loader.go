package main

// parses food data central csv into couchbase documents
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
func load(p ingest.Ingest) error {
	err := p.ProcessFiles(*i, dc)
	return err
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
		cs  fdc.Config
		err error
	)
	cs.GetConfig(c)
	dc.Conn = b
	// connect to datastore
	if err := dc.ConnectDs(cs); err != nil {
		log.Fatalln("Cannot connect to cluster ", err)
	}
	if dtype == fdc.BFPD {
		b := bfpd.Bfpd{Doctype: dt.ToString(fdc.BFPD)}
		err = load(b)
	} else if dtype == fdc.FNDDS {
		b := fndds.Fndds{Doctype: dt.ToString(fdc.FNDDS)}
		err = load(b)
	} else {
		b := dictionaries.Dictionary{Dt: dtype}
		err = load(b)
	}
	if err != nil {

	}

	log.Println("Finished.")
	dc.CloseDs()
	os.Exit(0)
}
