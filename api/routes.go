package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	gocb "gopkg.in/couchbase/gocb.v1"
	"gopkg.in/couchbase/gocb.v1/cbft"

	"github.com/littlebunch/gnut-bfpd-api/model"
)

// foodFdcID returns a single food based on a key value constructed from the fdcid
// If the format parameter equals 'meta' then only the food's meta-data is returned.
func foodFdcID(c *gin.Context) {
	var (
		q string
	)
	q = fmt.Sprintf("BFPD:%s", c.Param("id"))
	if c.Query("format") == META {
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
		if c.Query("format") == SERVING {
			c.JSON(http.StatusOK, f.Servings)
		} else if c.Query("format") == NUTRIENTS {
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
	if format == "" {
		format = META
	}
	if format != FULL && format != META && format != SERVING && format != NUTRIENTS {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("valid formats are %s or %s", META)})
		return
	}
	sort = c.Query("sort")
	if sort == "" {
		sort = "fdcId"
	}
	if sort != "" && sort != "foodDescription" && sort != "company" && sort != "fdcId" {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Unrecognized sort parameter.  Must be 'company', 'name' or 'fdcId'"})
		return
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

	q := fmt.Sprintf("select * from %s as gd where %s != '' offset %d limit %d", cs.CouchDb.Bucket, sort, offset, max)
	query := gocb.NewN1qlQuery(q)
	rows, err := b.ExecuteN1qlQuery(query, nil)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "No foods found"})
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
func foodsSearch(c *gin.Context) {
	var (
		max    int
		page   int
		format string
		foods  []interface{}
		query  *gocb.SearchQuery
	)
	// check the format parameter which defaults to BRIEF if not set
	format = c.Query("format")
	if format == "" {
		format = META
	}
	if format != FULL && format != META && format != SERVING && format != NUTRIENTS {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("valid formats are %s, %s, %s or %s", META, FULL, SERVING, NUTRIENTS)})
		return
	}
	max, err = strconv.Atoi(c.Query("max"))
	if err != nil {
		max = defaultListMax
	}
	if max > maxListSize {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("max parameter %d exceeds maximum allowed size of %d", max, maxListSize)})
		return
	}
	page, err = strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 0
	}
	if page < 0 {
		page = 0
	}
	offset := page * max
	indexName := cs.CouchDb.FtsIndex
	q := c.Query("q")
	f := c.Query("f")
	if f == "" {
		query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q)).Limit(int(max)).Skip(offset)
	} else {
		if f != "foodDescription" && f != "company" && f != "ingredients" {
			errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Unrecognized search field.  Must be one of 'foodDescription','company' or 'ingredients'"})
			return
		}
		query = gocb.NewSearchQuery(indexName, cbft.NewMatchQuery(q).Field(f)).Limit(int(max)).Skip(offset)
	}
	result, err := b.ExecuteSearchQuery(query)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Search query failed %v", err)})
		return
	}
	count := result.TotalHits()
	if format == META {
		var f fdc.FoodMeta
		for _, r := range result.Hits() {
			_, err := b.Get(r.Id, &f)
			if err != nil {
				errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
			} else {
				foods = append(foods, f)
			}

		}
	} else {
		var f fdc.Food
		for _, r := range result.Hits() {
			_, err := b.Get(r.Id, &f)
			if err != nil {
				errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
			} else {
				if format == SERVING {
					foods = append(foods, fdc.BrowseServings{FdcID: f.FdcID, Servings: f.Servings})
				} else if format == NUTRIENTS {
					foods = append(foods, fdc.BrowseNutrients{FdcID: f.FdcID, Nutrients: f.Nutrients})
				} else {
					foods = append(foods, f)
				}
				f = fdc.Food{}
			}

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
