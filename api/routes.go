package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	fdc "github.com/littlebunch/fdc-api/model"
)

func countsGet(c *gin.Context) {
	var counts []interface{}
	t := c.Param("doctype")
	if t == "" {
		if t = c.Query("doctype"); t == "" {
			errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Datasource is required!"})
			return
		}
	}
	if err := dc.Counts(cs.CouchDb.Bucket, t, &counts); err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No counts found!"})
		return
	}
	if counts != nil {
		c.JSON(http.StatusOK, counts[0])
	} else {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No counts found!"})

	}
	return
}

// foodFdcID returns a single food based on a key value constructed from the fdcid
// If the format parameter equals 'meta' then only the food's meta-data is returned.
func foodFdcID(c *gin.Context) {
	q := c.Param("id")
	if q == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "a FDC id in the q parameter is required"})
		return
	}
	var f fdc.Food
	err := dc.Get(q, &f)
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
	}
	if c.Query("format") == fdc.SERVING {
		c.JSON(http.StatusOK, f.Servings)
	} else {
		c.JSON(http.StatusOK, f)
	}
	return
}
func nutrientFdcID(c *gin.Context) {
	var q, n string

	if q = c.Param("id"); q == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "a FDC id in the q parameter is required"})
		return
	}
	if n = c.Query("n"); n == "" {
		var nd []interface{}
		q := fmt.Sprintf("select * from %s as nutrients where type='NUTDATA' and fdcId=\"%s\"", cs.CouchDb.Bucket, q)
		dc.Query(q, &nd)
		results := fdc.BrowseResult{Count: int32(len(nd)), Start: 0, Max: int32(len(nd)), Items: nd}
		c.JSON(http.StatusOK, results)

	} else {
		var nd fdc.NutrientData
		id := fmt.Sprintf("%s_%s", q, n)
		if err := dc.Get(id, &nd); err != nil {
			errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No nutrient data found!"})
		}
		c.JSON(http.StatusOK, nd)
	}

	return
}

// nutrientsBrowse returns the nutrients list
func nutrientsBrowse(c *gin.Context) {
	var (
		dt        fdc.DocType
	)
	max := 300
	page := 0
	nutrients,err := dc.GetDictionary(cs.CouchDb.Bucket, dt.ToString(fdc.NUT), int64(page), int64(max))
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Error."})
		return
	}
	results := fdc.BrowseResult{Count: int32(len(nutrients)), Start: int32(page), Max: int32(max), Items: nutrients}
	c.JSON(http.StatusOK, results)
}

// foodsBrowse returns a BrowseResult
func foodsBrowse(c *gin.Context) {
	var (
		max, page   int64
		sort, order string
		dt          fdc.DocType
	)
	if sort = c.Query("sort"); sort == "" {
		sort = "fdcId"
	}
	if sort != "" && sort != "foodDescription" && sort != "company" && sort != "fdcId" {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Unrecognized sort parameter.  Must be 'company', 'name' or 'fdcId'"})
		return
	}
	order, err := sortOrder(c.Query("order"))
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Unrecognized order parameter.  Must be 'asc' or 'desc'"})
		return
	}

	source := c.Query("source")
	if source != "" && dt.ToDocType(source) == 999 {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("Unrecognized source parameter.  Must be %s, %s or %s", dt.ToString(fdc.BFPD), dt.ToString(fdc.SR), dt.ToString(fdc.FNDDS))})
		return
	}
	if max, err = strconv.ParseInt(c.Query("max"), 10, 32); err != nil {
		max = defaultListMax
	}
	if max > maxListSize {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("max parameter %d exceeds maximum allowed size of %d", max, maxListSize)})
		return
	}
	if page, err = strconv.ParseInt(c.Query("page"), 10, 32); err != nil {
		page = 0
	}
	if page < 0 {
		page = 0
	}
	offset := page * max
	where := fmt.Sprintf("type=\"%s\" ",dt.ToString(fdc.FOOD))
	if source != "" {
		where = where + sourceFilter(source)
	}
	foods,err := dc.Browse(cs.CouchDb.Bucket, where, offset, max, sort, order)
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("Query error %v", err)})
		return
	}
	
	results := fdc.BrowseResult{Count: int32(len(foods)), Start: int32(page), Max: int32(max), Items: foods}
	c.JSON(http.StatusOK, results)
}

// foodsSearch runs a search and returns a BrowseResult
func foodsSearch(c *gin.Context) {
	var (
		max   int
		page  int
		foods []interface{}
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

	if max, err = strconv.Atoi(c.Query("max")); err != nil {
		max = defaultListMax
	}
	if max > maxListSize {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("max parameter %d exceeds maximum allowed size of %d", max, maxListSize)})
		return
	}
	if page, err = strconv.Atoi(c.Query("page")); err != nil {
		page = 0
	}
	if page < 0 {
		page = 0
	}
	offset := page * max

	if count, err = dc.Search(fdc.SearchRequest{Query: q, IndexName: cs.CouchDb.Fts, Format: fdc.META, Max: max, Page: offset}, &foods); err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Search query failed %v", err)})
		return
	}
	results := fdc.BrowseResult{Count: int32(count), Start: int32(page), Max: int32(max), Items: foods}
	c.JSON(http.StatusOK, results)
}

// foodsSearch runs a search as a POST and returns a BrowseResult
func foodsSearchPost(c *gin.Context) {
	var (
		foods []interface{}
		sr    fdc.SearchRequest
	)
	count := 0

	// check for a query
	err = c.BindJSON(&sr)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": err})
		return
	}
	if sr.Query == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Search query is required."})
		return
	}
	// check the format parameter which defaults to BRIEF if not set
	if sr.Format == "" {
		sr.Format = fdc.META
	} else if sr.Format != fdc.FULL && sr.Format != fdc.SERVING && sr.Format != fdc.NUTRIENTS {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("valid formats are %s, %s, %s or %s", fdc.META, fdc.FULL, fdc.SERVING, fdc.NUTRIENTS)})
		return
	}
	if &sr.Max == nil {
		sr.Max = defaultListMax
	} else if sr.Max > maxListSize || sr.Max < 0 {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("max parameter %d must be > 0 or <=  %d", sr.Max, maxListSize)})
		return
	}
	if &sr.Page == nil {
		sr.Page = 0
	}
	if sr.Page < 0 {
		sr.Page = 0
	}
	sr.Page = sr.Page * sr.Max
	sr.IndexName = cs.CouchDb.Fts
	if count, err = dc.Search(sr, &foods); err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Search query failed %v", err)})
		return
	}
	results := fdc.BrowseResult{Count: int32(count), Start: int32(sr.Page), Max: int32(sr.Max), Items: foods}
	c.JSON(http.StatusOK, results)
}

// nutrientReportPost produces a report of nutrient values and returns a BrowseResult
func nutrientReportPost(c *gin.Context) {
	var (
		nutdata []interface{}
		nr      fdc.NutrientReportRequest
		dt      *fdc.DocType
	)
	// check for a query
	err = c.BindJSON(&nr)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": err})
		return
	}
	if nr.Max <= 0 {
		nr.Max = defaultListMax
	} else if nr.Max > maxListSize || nr.Max < 0 {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("max parameter %d must be > 0 or <=  %d", nr.Max, maxListSize)})
		return
	}
	if &nr.Page == nil {
		nr.Page = 0
	}
	if nr.Page < 0 {
		nr.Page = 0
	}
	// validate values
	if nr.ValueLTE < 0 || nr.ValueGTE < 0 {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "ValueGTE  and ValueLTE must be greater than or equal to 0"})
		return
	} else if nr.ValueGTE == 0 && nr.ValueLTE == 0 {
		nr.ValueGTE = 0
		nr.ValueLTE = 100000
	} else if nr.ValueGTE > nr.ValueLTE {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("ValueGTE %d must be greater than or equal to ValueLTE  %d", nr.ValueGTE, nr.ValueLTE)})
		return
	}
	nr.Page = nr.Page * nr.Max
	// Datasource filter
	if nr.Source != "" && (dt.ToString(fdc.FNDDS) != nr.Source && dt.ToString(fdc.SR) != nr.Source && dt.ToString(fdc.BFPD) != nr.Source) {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Source must be one of BFPD, FNDDS, or SR."})
		return
	}
	if err = dc.NutrientReport(cs.CouchDb.Bucket, nr, &nutdata); err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Data error %v", err)})
		return
	}
	results := fdc.BrowseResult{Count: int32(len(nutdata)), Start: int32(nr.Page), Max: int32(nr.Max), Items: nutdata}
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
func sourceFilter(s string) string {
	w := ""
	if s != "" {
		if s == "BFPD" {
			w = fmt.Sprintf(" AND ( dataSource = '%s' OR dataSource='%s' )", "LI", "GDSN")
		} else {
			w = fmt.Sprintf(" AND dataSource = '%s'", s)
		}
	}
	return w
}
func sortOrder(o string) (string, error) {
	order := o
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		return "", errors.New("Unrecognized order parameter.  Must be 'asc' or 'desc'")
	}
	return order, nil
}
