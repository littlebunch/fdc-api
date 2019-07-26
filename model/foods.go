// Package fdc describes food products data model
package fdc

import (
	"time"
)

// BrowseResult is returned from the browse endpoints
type BrowseResult struct {
	Count int32         `json:"count"`
	Start int32         `json:"start"`
	Max   int32         `json:"max"`
	Items []interface{} `json:"items"`
}

// BrowseServings is returned from the browse endpoints
type BrowseServings struct {
	FdcID    string    `json:"fdcId" binding:"required"`
	Servings []Serving `json:"servingSizes"`
}

// BrowseNutrients is returned from the browse endpoints
type BrowseNutrients struct {
	FdcID     string         `json:"fdcId" binding:"required"`
	Nutrients []NutrientData `json:"nutrients"`
}

// NutrientRequest wraps a POST nutrient report
type NutrientReportRequest struct {
	Page     int `json:"page"`
	Max      int `json:"max"`
	Nutrient int `json:"nutrientno" binding:"required"`
	ValueGTE int `json:"valueGte"`
	ValueLTE int `json:"valueLte`
}

// SearchRequest wraps a POST search
type SearchRequest struct {
	Query       string `json:"q" binding:"required"`
	Format      string `json:"format,omitEmpty"`
	SearchField string `json:"searchfield,omitEmpty"`
	Page        int    `json:"page"`
	Max         int    `json:"max"`
	Sort        string `json:"sort,omitEmpty"`
	SearchType  string `json:"searchtype,omitEmpty"`
	IndexName   string `json:"indexname"`
}

// SearchResult is returned from the search endpoints
type SearchResult struct {
	Food      FoodMeta       `json:"foodMeta"`
	Servings  []Serving      `json:"servingSizes"`
	Nutrients []NutrientData `json:"nutrients"`
}

// SearchServings is returned from the search endpoints when format=servings
type SearchServings struct {
	Food     FoodMeta  `json:"foodMeta"`
	Servings []Serving `json:"servingSizes"`
}

// SearchNutrients is returned from the search endpoints when format=nutrients
type SearchNutrients struct {
	Food      FoodMeta       `json:"foodMeta"`
	Nutrients []NutrientData `json:"nutrients"`
}

// FoodMeta abbreviated Food containing only meta-data
type FoodMeta struct {
	FdcID        string `json:"fdcId" binding:"required"`
	Upc          string `json:"upc"`
	Description  string `json:"foodDescription" binding:"required"`
	Ingredients  string `json:"ingredients,omitempty"`
	Source       string `json:"dataSource"`
	Manufacturer string `json:"company,omitempty"`
	Type         string `json:"type"`
}

// Food reflects JSON used to transfer BFPD foods data from USDA csv
type Food struct {
	UpdatedAt       time.Time  `json:"lastChangeDateTime,omitempty"`
	FdcID           string     `json:"fdcId" binding:"required"`
	NdbNo           string     `json:"ndbno,omitempty"`
	Upc             string     `json:"upc,omitempty"`
	Description     string     `json:"foodDescription" binding:"required"`
	Source          string     `json:"dataSource"`
	PublicationDate time.Time  `json:"publicationDateTime"`
	Ingredients     string     `json:"ingredients,omitempty"`
	Manufacturer    string     `json:"company,omitempty"`
	Group           *FoodGroup `json:"foodGroup,omitempty"`
	Servings        []Serving  `json:"servingSizes,omitempty"`
	//Nutrients       []NutrientData `json:"nutrients,omitempty"`
	Type       string      `json:"type" binding:"required"`
	InputFoods []InputFood `json:"inputfoods,omitempty"`
}

// InputFood describes an FNDDS Input Food
type InputFood struct {
	Description        string  `json:"foodDescription" binding:"required"`
	SeqNo              int     `json:"seq"`
	Amount             float32 `json:"amount"`
	SrCode             int     `json:"srcode,omitempty"`
	Unit               string  `json:"unit"`
	Portion            string  `json:"portion,omitempty"`
	PortionDescription string  `json:"portionDescription,omitempty"`
	Weight             float32 `json:"weight"`
}

// Serving describes a list nutrients for a given state, weight and amount
// A subdocument of Food
type Serving struct {
	Nutrientbasis string  `json:"100UnitNutrientBasis,omitempty"`
	Description   string  `json:"householdServingUom"`
	Servingstate  string  `json:"servingState,omitempty"`
	Weight        float32 `json:"weightInGmOrMl"`
	Servingamount float32 `json:"householdServingValue,omitempty"`
	Datapoints    int32   `json:"datapoints,omitempty"`
}

// Nutrient is metadata abount nutrients usually in a nutrients collection
type Nutrient struct {
	NutrientID uint   `json:"id" binding:"required"`
	Nutrientno uint   `json:"nutrientno" binding:"required"`
	Tagname    string `json:"tagname,omitempty"`
	Name       string `json:"name"  binding:"required"`
	Unit       string `json:"unit"  binding:"required"`
	Type       string `json:"type"  binding:"required"`
}

// Derivation is a code for describing how nutrient values are derived
// A subdocument of NutrientData
type Derivation struct {
	ID          int32  `json:"id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// NutrientData is the list of nutrient values
// A document of type NUTDATA
type NutrientData struct {
	FdcID      string      `json:"fdcId" binding:"required"`
	Type       string      `json:"type"`
	Value      float32     `json:"valuePer100UnitServing"`
	Unit       string      `json:"unit"  binding:"required"`
	Derivation *Derivation `json:"derivation,omitempty"`
	Nutrientno uint        `json:"nutrientNumber"`
	Nutrient   string      `json:"nutrientName"`
	Datapoints int         `json:"datapoints,omitempty"`
	Min        float32     `json:"min,omitempty"`
	Max        float32     `json:"max,omitempty"`
}

// NutrientDataValue
type NutrientDataValue struct {
	FdcID      string  `json:"fdcId" binding:"required"`
	Value      float32 `json:"valuePer100UnitServing"`
	Nutrientno uint    `json:"nutrientNumber"`
	Type       string  `json:"type"`
}

// FoodGroup is the dictionary of FNDDS and SR food groups
type FoodGroup struct {
	ID          int32  `json:"id" binding:"required"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description" binding:"required"`
	LastUpdate  string `json:"lastUpdate,omitempty"`
	Type        string `json:"type" binding:"required"`
}
