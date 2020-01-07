package main

//test gin request handlers
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	fdc "github.com/littlebunch/fdc-api/model"
)

type rp struct {
	Status int    `json:"status"`
	Data   string `json:"data"`
}

func TestFoodGetHandler(t *testing.T) {
	handler := func(c *gin.Context) {
		if c.Param("id") == "1" {
			c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": `{"fdcId":"1","desc":"Test Food","company":"Test Company}`})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "data": nil})
		}

	}
	var parsed map[string]interface{}
	router := gin.New()
	router.GET("/food/:id", handler)

	req, _ := http.NewRequest("GET", "/food/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	err := json.Unmarshal(resp.Body.Bytes(), &parsed)
	if err != nil {
		t.Errorf("error is %v", err)
	} else if parsed["status"] != float64(http.StatusOK) {
		t.Errorf("Expecting %d status is %d", http.StatusOK, parsed["status"])
	}
}
func TestNutrientsGetHandler(t *testing.T) {
	handler := func(c *gin.Context) {
		_nutrients := `{[{"nutrient":"Test Nutrient1"},{"nutrient":"Test Nutrient2"}]}`
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _nutrients})
	}
	var parsed map[string]interface{}
	router := gin.New()
	router.GET("/nutrients/browse", handler)

	req, _ := http.NewRequest("GET", "/nutrients/browse", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	err := json.Unmarshal(resp.Body.Bytes(), &parsed)
	if err != nil {
		t.Errorf("error is %v", err)
	} else if parsed["status"] != float64(http.StatusOK) {
		t.Errorf("Expecting %d status is %d", http.StatusOK, parsed["status"])
	}
}
func TestFoodsBrowseGetHandler(t *testing.T) {
	handler := func(c *gin.Context) {
		page, err := strconv.ParseInt(c.Query("page"), 10, 32)
		if err != nil || int32(page) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid Page Parameter SB 1"})
		} else if c.Query("sort") != "company" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid Sort Parameter SB 'company'"})
		} else {
			_foods := `{[{"fdcId":"1","desc":"Test Food1"},{"fdcId":"2","desc":"Test Food2"}]}`
			c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _foods})
		}
	}
	var parsed map[string]interface{}
	router := gin.New()
	router.GET("/foods/browse", handler)

	req, _ := http.NewRequest("GET", "/foods/browse?page=1&max=100&sort=company", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	err := json.Unmarshal(resp.Body.Bytes(), &parsed)
	if err != nil {
		t.Errorf("error is %v", err)
	} else if parsed["status"] != float64(http.StatusOK) {
		t.Errorf("Expecting %d status is %d %s", http.StatusOK, parsed["status"], parsed["message"])
	}
}
func TestFoodsSearchPostHandler(t *testing.T) {
	handler := func(c *gin.Context) {
		var sr fdc.SearchRequest
		err := c.BindJSON(&sr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": fmt.Sprintf("Invalid JSON in request: %v", err)})
		} else if sr.Query != "broccoli" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid query SB 'broccoli'"})
		} else {
			var _items []interface{}
			_items = append(_items, `{"id":"1","desc":"Test Food1"}`)
			_items = append(_items, `{"id":"2","desc":"Test Food2"}`)
			c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": fdc.BrowseResult{Count: 2, Start: 1, Max: 2, Items: _items}})
		}
	}
	var parsed map[string]interface{}
	router := gin.New()
	router.POST("/foods/search", handler)

	req, _ := http.NewRequest("POST", "/foods/search", strings.NewReader(`{"q":"broccoli","searchfield":"foodDescription","max":50,"page":0}`))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	err := json.Unmarshal(resp.Body.Bytes(), &parsed)
	if err != nil {
		t.Errorf("error is %v", err)
	} else if parsed["status"] != float64(http.StatusOK) {
		t.Errorf("Expecting %d status is %d message is %s", http.StatusOK, parsed["status"], parsed["message"])
	}
}
