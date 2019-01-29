package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/littlebunch/gnutdata-bfpd-api/model"
)

// foodFdcID returns a single food based on a key value constructed from the fdcid
// If the format parameter equals 'meta' then only the food's meta-data is returned.
func foodFdcID(c *gin.Context) {
	var q string
	q = fmt.Sprintf("BFPD:%s", c.Param("id"))
	if c.Query("format") == fdc.META {
		var f fdc.FoodMeta
		err := dc.Get(q, &f)
		if err != nil {
			errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
		} else {
			c.JSON(http.StatusOK, f)
		}
	} else {
		var f fdc.Food
		err := dc.Get(q, &f)
		if err != nil {
			errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
		}
		if c.Query("format") == fdc.SERVING {
			c.JSON(http.StatusOK, f.Servings)
		} else if c.Query("format") == fdc.NUTRIENTS {
			c.JSON(http.StatusOK, f.Nutrients)
		} else {
			c.JSON(http.StatusOK, f)
		}
	}

	return
}

// foodsGet returns a BrowseResult
func foodsGet(c *gin.Context) {
	var (
		max    int64
		page   int64
		count  int
		format string
		sort   string
		foods  []interface{}
	)
	// check the format parameter which defaults to META if not set
	format = c.Query("format")
	if format == "" {
		format = fdc.META
	}
	if format != fdc.FULL && format != fdc.META && format != fdc.SERVING && format != fdc.NUTRIENTS {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("valid formats are %s, %s, %s or %s", fdc.META, fdc.FULL, fdc.SERVING, fdc.NUTRIENTS)})
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
	dc.Browse(cs.CouchDb.Bucket, offset, max, format, sort, &foods)
	results := fdc.BrowseResult{Count: int32(count), Start: int32(offset), Max: int32(max), Items: foods}
	c.JSON(http.StatusOK, results)
}

// foodsSearch runs a search and returns a BrowseResult
func foodsSearch(c *gin.Context) {
	var (
		max    int
		page   int
		format string
		foods  []interface{}
	)
	count := 0
	// check for a query
	q := c.Query("q")
	if q == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "A search string in the q parameter is required"})
		return
	}
	// check for field
	f := c.Query("f")
	if f != "" && f != "foodDescription" && f != "upc" && f != "company" && f != "ingredients" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Unrecognized search field.  Must be one of 'foodDescription','company', 'upc' or 'ingredients'"})
		return
	}
	// check the format parameter which defaults to BRIEF if not set
	format = c.Query("format")
	if format == "" {
		format = fdc.META
	}
	if format != fdc.FULL && format != fdc.META && format != fdc.SERVING && format != fdc.NUTRIENTS {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("valid formats are %s, %s, %s or %s", fdc.META, fdc.FULL, fdc.SERVING, fdc.NUTRIENTS)})
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

	count, err = dc.Search(q, f, cs.CouchDb.FtsIndex, format, max, offset, &foods)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Search query failed %v", err)})
		return
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
