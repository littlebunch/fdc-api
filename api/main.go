package main

// @APITitle Brand Foods Product Database

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/littlebunch/gnutdata-bfpd-api/ds"
	"github.com/littlebunch/gnutdata-bfpd-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
)

const (
	maxListSize    = 150
	defaultListMax = 50
	apiVersion     = "1.0.0 Beta"
)

var (
	d   = flag.Bool("d", false, "Debug")
	i   = flag.Bool("i", false, "Initialize the authentication store")
	c   = flag.String("c", "config.yml", "YAML Config file")
	l   = flag.String("l", "/tmp/bfpd.out", "send log output to this file -- defaults to /tmp/bfpd.out")
	p   = flag.String("p", "8000", "TCP port to used")
	r   = flag.String("r", "v1", "root path to deploy -- defaults to 'v1'")
	b   *gocb.Bucket
	cs  fdc.Config
	dc  ds.DS
	err error
)

// process cli flags; build the config and init an Mongo client and a logger
func init() {
	var (
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
	flag.Parse()
	// get configuration
	cs.GetConfig(c)
	dc.Conn = b
	// Connect to couchbase datastore
	err = dc.ConnectDs(cs)
	if err != nil {
		log.Fatal("Cannot get datastore connection.")
	}
	defer dc.CloseDs()
	// initialize our jwt authentication
	//var u *auth.User
	//if *i {
	//	u.BootstrapUsers(session, cs.MongoDb.Collection)
	//}
	//authMiddleware := u.AuthMiddleware(session, cs.MongoDb.Collection)
	//router := gin.Default()
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	v1 := router.Group(fmt.Sprintf("%s", *r))
	{
		//v1.POST("/login", authMiddleware.LoginHandler)
		v1.GET("/food/:id/:format", foodFdcID)
		v1.GET("/food/:id", foodFdcID)
		v1.GET("/browse", foodsGet)
		v1.GET("/search", foodsSearch)
		v1.GET("/count/:doctype", countsGet)
		//v1.POST("/user/", authMiddleware.MiddlewareFunc(), userPost)
	}
	endless.ListenAndServe(":"+*p, router)

}
