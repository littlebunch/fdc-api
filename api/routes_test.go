package main

//test gin request handlers
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type rp struct {
	Status int    `json:"status"`
	Data   string `json:"data"`
}

func TestFoodGetHandler(t *testing.T) {
	handler := func(c *gin.Context) {
		//c.String(http.StatusOK, "bar")
		_food := `{"food":"Test Food"}`
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _food})
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
