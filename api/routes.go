package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	fdc "github.com/littlebunch/fdc-api/model"
)

var isUpc = regexp.MustCompile(`^[0-9]+$`)

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

// foodFdcID returns a single food in a BrowseResult based on a key value constructed from the fdcId
// or upc.  Any id that looks like a upc gets converted to a fdcId
func foodFdcID(c *gin.Context) {
	var (
		f     fdc.Food
		items []interface{}
	)
	q := c.Param("id")
	if q == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "a FDC id in the q parameter is required"})
		return
	}
	// convert anything that looks a upc to an fdcId
	if len(q) > 7 {
		q, _ = upcTofdcid(q, cs.CouchDb.Bucket)
	}
	err := dc.Get(q, &f)
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No food found!"})
	}
	items = append(items, f)
	results := fdc.BrowseResult{Count: 1, Start: 0, Max: 1, Items: items}
	c.JSON(http.StatusOK, results)
	return
}

// returns foods in a BrowseResult for a list of fdcIds or upcs.  If an id looks like a upc it is converted
// to a fdcId.
func foodFdcIds(c *gin.Context) {
	var (
		dt fdc.DocType
		f  []interface{}
	)
	qids, err := buildIDList(getFdcIDs(c.QueryArray("id")))
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Cannot request more than 24 id's"})
		return
	}
	q := fmt.Sprintf("SELECT * from %s WHERE type=\"%s\" AND fdcId in %s", cs.CouchDb.Bucket, dt.ToString(fdc.FOOD), qids)
	dc.Query(q, &f)
	results := fdc.BrowseResult{Count: int32(len(f)), Start: 0, Max: int32(len(f)), Items: f}
	c.JSON(http.StatusOK, results)

	return

}

// returns a dictionary list which can be nutrients (NUT), derivations (DERV), food categories (FGGPC)
func dictionaryBrowse(c *gin.Context) {
	var (
		dt        fdc.DocType
		t         string
		max, page int64
	)
	t = c.Param("type")
	if t == "" {
		t = dt.ToString(fdc.NUT)
	}
	if t != "NUT" && t != "DERV" && t != "FGSR" && t != "FGFNDDS" && t != "FGGPC" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "one of type parameter is required: NUT, DERV, FGSR,FGFNDDS, FGGPC"})
		return
	}
	if max, err = strconv.ParseInt(c.Query("max"), 10, 32); err != nil {
		max = 300
	}
	if page, err = strconv.ParseInt(c.Query("page"), 10, 32); err != nil {
		page = 0
	}
	if page < 0 {
		page = 0
	}
	offset := page * max
	items, err := dc.GetDictionary(cs.CouchDb.Bucket, t, offset, max)
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Error."})
		return
	}

	results := fdc.BrowseResult{Count: int32(len(items)), Start: int32(offset), Max: int32(max), Items: items}
	c.JSON(http.StatusOK, results)
}

// returns nutrients for a specified foods identified by fdcId or UPC
// if an optional n parameter is provided then limit nutrients returned to the
// nutrientno's in the n paramter array
func nutrientFdcID(c *gin.Context) {
	var (
		q   string
		dt  fdc.DocType
		nd  []interface{}
		ndb []fdc.NutrientFoodBrowseItem
		ndi fdc.NutrientFoodBrowseItem
		ndd fdc.NutrientFoodBrowse
	)

	if q = c.Param("id"); q == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "a FDC id in the q parameter is required"})
		return
	}
	// replace UPC with fdcId
	if len(q) > 7 {
		q, _ = upcTofdcid(q, cs.CouchDb.Bucket)
	}
	// build query for one or more nutrient #'s otherwise build a query to return all nutrients
	if n := c.QueryArray("n"); len(n) > 0 {

		var nids []string
		for i := range n {
			nids = append(nids, fmt.Sprintf("%s_%s", q, n[i]))
		}
		qids, _ := buildIDList(nids)
		q = fmt.Sprintf("SELECT fdcId,upc,foodDescription,company,category,valuePer100UnitServing,unit,nutrientNumber,nutrientName from %s as nutrient WHERE type=\"%s\" AND meta(nutrient).id in %s", cs.CouchDb.Bucket, dt.ToString(fdc.NUTDATA), qids)
	} else {

		q = fmt.Sprintf("SELECT fdcId,upc,foodDescription,company,category,valuePer100UnitServing,unit,nutrientNumber,nutrientName from %s as nutrient WHERE type=\"%s\" AND fdcId = \"%s\"", cs.CouchDb.Bucket, dt.ToString(fdc.NUTDATA), q)
	}
	dc.Query(q, &nd)
	haveFood := false
	for i := range nd {
		b, _ := json.Marshal(nd[i])
		json.Unmarshal(b, &ndi)
		if !haveFood {
			json.Unmarshal(b, &ndd)
			haveFood = true
		}

		ndb = append(ndb, ndi)
	}
	results := fdc.NutrientFoodBrowse{FdcID: ndd.FdcID, Description: ndd.Description, Upc: ndd.Upc, Nutrients: ndb}
	//results := fdc.BrowseResult{Count: int32(len(nd)), Start: 0, Max: int32(len(nd)), Items: nd}
	c.JSON(http.StatusOK, results)

	return
}

// returns nutrients for a specified list of foods identified by fdcId
// if an optional n parameter is provided then limit nutrients returned to the
// nutrientno in the n paramter
func nutrientFdcIDs(c *gin.Context) {
	var (
		q       string
		dt      fdc.DocType
		nd      []interface{}
		ndb     []fdc.NutrientFoodBrowseItem
		ndi     fdc.NutrientFoodBrowseItem
		ndd, nn fdc.NutrientFoodBrowse
		ndds    []fdc.NutrientFoodBrowse
	)
	// replace any UPC's with FdcID's
	ids := getFdcIDs(c.QueryArray("id"))
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Cannot request more than 24 id's"})
		return
	}
	// create nutrient data ids
	if n := c.QueryArray("n"); len(n) > 0 {
		var nids []string
		for id := range ids {
			for i := range n {
				nids = append(nids, fmt.Sprintf("%s_%s", ids[id], n[i]))
			}
		}
		qids, _ := buildIDList(nids)
		q = fmt.Sprintf("SELECT fdcId,upc,foodDescription,company,category,derivation,valuePer100UnitServing,unit,nutrientNumber,nutrientName from %s as nutrient WHERE type=\"%s\" AND meta(nutrient).id in %s order by fdcId", cs.CouchDb.Bucket, dt.ToString(fdc.NUTDATA), qids)

	} else {
		qids, _ := buildIDList(ids)
		q = fmt.Sprintf("SELECT fdcId,upc,foodDescription,company,category,derivation,valuePer100UnitServing,unit,nutrientNumber,nutrientName from %s as nutrient WHERE type=\"%s\" AND fdcId in %s order by fdcId", cs.CouchDb.Bucket, dt.ToString(fdc.NUTDATA), qids)
	}
	fmt.Printf("q=%s\n", q)
	dc.Query(q, &nd)
	haveFood := false
	for i := range nd {
		b, _ := json.Marshal(nd[i])
		json.Unmarshal(b, &nn)
		json.Unmarshal(b, &ndi)
		if haveFood && nn.FdcID != ndd.FdcID {
			ndd.Nutrients = ndb
			ndds = append(ndds, ndd)
			ndb = nil
			haveFood = false
		} else {
			if !haveFood {
				ndd.Category = nn.Category
				ndd.Description = nn.Description
				ndd.FdcID = nn.FdcID
				ndd.Manufacturer = nn.Manufacturer
				ndd.Upc = nn.Upc
				haveFood = true
			}

		}
		ndb = append(ndb, ndi)

	}
	ndd.Nutrients = ndb
	ndds = append(ndds, ndd)
	c.JSON(http.StatusOK, ndds)
	return
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
	where := fmt.Sprintf("type=\"%s\" ", dt.ToString(fdc.FOOD))
	// Check for filter on food group description or id.  Add to query if present
	if fg := c.Query("fg"); fg != "" {
		if i, err := strconv.ParseInt(fg, 0, 32); err == nil {
			where += fmt.Sprintf(" AND foodGroup.id=%d", i)
		} else {
			where += fmt.Sprintf(" AND foodGroup.description=\"%s\"", fg)
		}

	}
	if source != "" {
		where = where + sourceFilter(source)
	}
	foods, err := dc.Browse(cs.CouchDb.Bucket, where, offset, max, sort, order)
	if err != nil {
		errorout(c, http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": fmt.Sprintf("Query error %v", err)})
		return
	}
	results := fdc.BrowseResult{Count: int32(len(foods)), Start: int32(page), Max: int32(max), Items: foods}
	c.JSON(http.StatusOK, results)
}

// foodsSearch runs a simple keyword search and returns a BrowseResult
func foodsSearchGet(c *gin.Context) {
	var (
		max, page int
		err       error
	)
	// check for a query
	q := c.Query("q")
	if q == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "A search string in the q parameter is required"})
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

	results, err := search(fdc.SearchRequest{Query: q, IndexName: cs.CouchDb.Fts, Max: max, Page: offset})
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Search query failed %v", err)})
		return
	}
	c.JSON(http.StatusOK, results)
}

// foodsSearch runs a SearchRequest as a POST and returns a BrowseResult
func foodsSearchPost(c *gin.Context) {
	var sr fdc.SearchRequest
	// check for a query
	err := c.BindJSON(&sr)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Invalid JSON in request: %v", err)})
		return
	}
	if sr.Query == "" {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Search query is required."})
		return
	}
	if sr.Max == 0 {
		sr.Max = defaultListMax
	} else if sr.Max > maxListSize || sr.Max < 0 {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("max parameter %d must be > 0 or <=  %d", sr.Max, maxListSize)})
		return
	}
	if sr.Page < 0 {
		sr.Page = 0
	}
	// only run REGEX searches against a keyword index
	if sr.SearchType == fdc.REGEX {
		sr.SearchField += "_kw"
	}
	sr.Page = sr.Page * sr.Max
	sr.IndexName = cs.CouchDb.Fts
	results, err := search(sr)
	if err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Search query failed %v", err)})
		return
	}
	c.JSON(http.StatusOK, results)
}

// returns openapi spec in either json or yaml format
func specDoc(c *gin.Context) {
	t := c.Param("type")
	if t == "" || (t != "yaml" && t != "json") {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "a doc type is required: yaml or json"})
		return
	}
	if t == "yaml" {
		raw, err := ioutil.ReadFile(YAMLSPEC)
		if err != nil {
			log.Println(err.Error())
			errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Cannot retrieve YAML doc"})
			return
		}
		c.Data(http.StatusOK, gin.MIMEYAML, raw)
	} else {
		raw, err := ioutil.ReadFile(JSONSPEC)
		if err != nil {
			log.Println(err.Error())
			errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Cannot retrieve JSON doc"})
			return
		}
		c.Data(http.StatusOK, gin.MIMEJSON, raw)
	}
	return
}

// search performs a SearchRequest on a datastore search and returns the result
func search(sr fdc.SearchRequest) (fdc.BrowseResult, error) {
	var (
		r   []interface{}
		err error
	)
	count := 0
	if count, err = dc.Search(sr, &r); err != nil {
		return fdc.BrowseResult{}, err
	}
	results := fdc.BrowseResult{Count: int32(count), Start: int32(sr.Page), Max: int32(sr.Max), Items: r}
	return results, nil
}

// nutrientReportPost produces a report of nutrient values and returns a BrowseResult
func nutrientReportPost(c *gin.Context) {
	var (
		nutdata []interface{}
		nr      fdc.NutrientReportRequest
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
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("ValueGTE %f must be greater than or equal to ValueLTE  %f", nr.ValueGTE, nr.ValueLTE)})
		return
	}
	nr.Page = nr.Page * nr.Max

	if err = dc.NutrientReport(cs.CouchDb.Bucket, nr, &nutdata); err != nil {
		errorout(c, http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Data error %v", err)})
		return
	}
	results := fdc.BrowseNutrientReport{Request: nr, Items: nutdata}
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

// converts an array of ids to a query string of the form ["12345",23456",...]
func buildIDList(ids []string) (string, error) {
	var (
		err  error
		qids string
	)
	if len(ids) > 24 {
		err = errors.New("Cannot request more than 24 id's")
	} else {
		qids = "["
		for id := range ids {
			qids += fmt.Sprintf("\"%s\",", ids[id])
		}
		qids = strings.Trim(qids, ",")
		qids += "]"
	}
	return qids, err
}

// convert UPC codes to fdc ids as necessary and return transformed array
func getFdcIDs(ids []string) []string {
	var (
		ids2 []string
		nid  string
	)
	for id := range ids {
		if len(ids[id]) > 7 && isUpc.MatchString(ids[id]) {
			nid, _ = upcTofdcid(ids[id], cs.CouchDb.Bucket)
			ids2 = append(ids2, nid)
		} else {
			ids2 = append(ids2, ids[id])
		}
	}
	return ids2
}

// return fdcId from UPC look-up
func upcTofdcid(upc string, bucket string) (string, error) {
	type f struct {
		FdcID string `json:"fdcId" binding:"required"`
	}
	var (
		r   []interface{}
		fid f
		j   []byte
	)
	q := fmt.Sprintf("SELECT fdcId from %s where upc = \"%s\" AND type=\"FOOD\"", bucket, upc)
	if err := dc.Query(q, &r); err != nil {
		log.Printf("%v\n", err)
		return "", err
	}
	for i := range r {
		if j, err = json.Marshal(r[i]); err != nil {
			log.Printf("%s %v %v\n", upc, j, err)
			return "", err
		}
		if err = json.Unmarshal(j, &fid); err != nil {
			log.Printf("%s %s %v\n", upc, string(j), err)
			return "", err
		}

	}
	return fid.FdcID, nil
}
