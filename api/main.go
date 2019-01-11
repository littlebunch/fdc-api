package main

// @APITitle Brand Foods Product Database

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/littlebunch/gnut-bfpd-api/model"
	gocb "gopkg.in/couchbase/gocb.v1"
)

const (
	maxListSize    = 150
	defaultListMax = 50
	apiVersion     = "1.0.0 Beta"
	META           = "meta"
	SERVING        = "servings"
	NUTRIENTS      = "nutrients"
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
	// Connect to couchbase
	cluster, err := gocb.Connect("couchbase://" + cs.CouchDb.URL)
	if err != nil {
		log.Fatalln("Cannot connect to cluster ", err)
	}
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: cs.CouchDb.User,
		Password: cs.CouchDb.Pwd,
	})
	b, err = cluster.OpenBucket(cs.CouchDb.Bucket, "")
	if err != nil {
		log.Fatalln("Cannot connect to bucket!", err)
	}
	defer b.Close()
	defer cluster.Close()
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
		//	v1.GET("/browse", foodsGet)
		v1.GET("/food/:id/:format", foodFdcID)
		v1.GET("/food/:id", foodFdcID)
		v1.GET("/foods", foodsGet)
		//	v1.GET("/nutrient/report", foodNutReportGet)
		//	v1.GET("/nutrient/list", nutListGet)
		//v1.POST("/user/", authMiddleware.MiddlewareFunc(), userPost)
	}
	endless.ListenAndServe(":"+*p, router)

}

// foodFdcID returns a single food based on a key value constructed from the fdcid
// If the format parameter equals 'meta' then only the food's meta-data is returned.
func foodFdcID(c *gin.Context) {
	var (
		q string
	)
	q = fmt.Sprintf("BFPD:%s", c.Param("id"))
	if c.Param("format") == META {
		var f fdc.FoodMeta
		_, err := b.Get(q, &f)
		if err != nil {
			errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
		} else {
			c.JSON(http.StatusOK, f)
		}
	} else {
		var f fdc.Food
		_, err := b.Get(q, &f)
		if err != nil {
			errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
		}
		if c.Param("format") == SERVING {
			c.JSON(http.StatusOK, f.Servings)
		} else if c.Param("format") == NUTRIENTS {
			c.JSON(http.StatusOK, f.Nutrients)
		} else {
			c.JSON(http.StatusOK, f)
		}
	}

	return
}

func foodsGet(c *gin.Context) {
	var (
		max    int64
		page   int64
		count  int
		format string
		sort   string
		foods  []interface{}
	)
	// check the format parameter which defaults to BRIEF if not set
	format = c.Query("format")
	if format != "" && (format != META && format != SERVING && format != NUTRIENTS) {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("valid formats are %s or %s", META)})
		return
	}
	if sort != "" && sort == "name" {
		sort = "foodDescription"
	} else {
		sort = "fdcId"
	}
	max, err = strconv.ParseInt(c.Query("max"), 10, 32)
	if err != nil {
		max = defaultListMax
	}
	if max > maxListSize {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("max parameter %d exceeds maximum allowed size of %d", max, maxListSize)})
		return
	}
	page, err = strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		page = 0
	}
	if page < 0 {
		page = 0
	}
	offset := page * max

	q := fmt.Sprintf("select * from %s as gd offset %d limit %d", cs.CouchDb.Bucket, offset, max)
	query := gocb.NewN1qlQuery(q)
	rows, err := b.ExecuteN1qlQuery(query, nil)
	if err != nil {
		fmt.Printf("ERROR=%v for %v", err, query)
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No foods found"})
	}
	if format == META {
		type n1q struct {
			Item fdc.FoodMeta `json:"gd"`
		}
		var row n1q
		for rows.Next(&row) {
			foods = append(foods, row.Item)
		}
	} else {
		type n1q struct {
			Item fdc.Food `json:"gd"`
		}
		var row n1q
		for rows.Next(&row) {
			if format == SERVING {
				foods = append(foods, fdc.BrowseServings{FdcID: row.Item.FdcID, Servings: row.Item.Servings})
			} else if format == NUTRIENTS {
				foods = append(foods, fdc.BrowseNutrients{FdcID: row.Item.FdcID, Nutrients: row.Item.Nutrients})
			} else {
				foods = append(foods, row.Item)
			}
			row = n1q{}
		}
	}
	results := fdc.BrowseResult{Count: int32(count), Start: int32(offset), Max: int32(max), Items: foods}
	c.JSON(http.StatusOK, results)
}

// errorout
func errorout(c *gin.Context, status int, data gin.H) {
	switch c.Request.Header.Get("Accept") {
	case "application/xml":
		c.XML(status, data)
	default:
		c.JSON(status, data)
	}
}
