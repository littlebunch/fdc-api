// Package fdc describes food products data model
package fdc

import (
	"time"
)

// FoodMeta abbreviated Food containing only meta-data
type FoodMeta struct {
	FdcID        string `json:"fdcId" binding:"required"`
	Upc          string `json:"upc"`
	Description  string `json:"foodDescription" binding:"required"`
	Ingredients  string `json:"ingredients,omitempty"`
	Source       string `json:"dataSource"`
	Manufacturer string `json:"company,omitempty"`
	Type         string `json:"type"`
	Category     string `json:"foodgroup.description,omitempty"`
}

// Food reflects JSON used to transfer BFPD foods data from USDA csv
type Food struct {
	ID              string      `json:"_id,omitempty"`
	Rev             string      `json:"_rev,omitempty"`
	UpdatedAt       time.Time   `json:"lastChangeDateTime,omitempty"`
	FdcID           string      `json:"fdcId" binding:"required"`
	NdbNo           string      `json:"ndbno,omitempty"`
	Upc             string      `json:"upc,omitempty"`
	Description     string      `json:"foodDescription" binding:"required"`
	Source          string      `json:"dataSource"`
	PublicationDate time.Time   `json:"publicationDateTime"`
	ModifiedDate    time.Time   `json:"modifiedDate,omitempty"`
	AvailableDate   time.Time   `json:"availableDate,omitempty"`
	DiscontinueDate time.Time   `json:"discontinueDate,omitempty"`
	Ingredients     string      `json:"ingredients,omitempty"`
	Manufacturer    string      `json:"company,omitempty"`
	Group           *FoodGroup  `json:"foodGroup,omitempty"`
	Servings        []Serving   `json:"servingSizes,omitempty"`
	Type            string      `json:"type" binding:"required"`
	Country         string      `json:"marketCountry,omitempty"`
	InputFoods      []InputFood `json:"inputfoods,omitempty"`
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
	Nutrientbasis string  `json:"nutrientBasis,omitempty"`
	Description   string  `json:"servingUnit"`
	Servingstate  string  `json:"servingState,omitempty"`
	Weight        float32 `json:"weight"`
	Servingamount float32 `json:"value,omitempty"`
	Datapoints    int32   `json:"dataPoints,omitempty"`
}

// Nutrient is metadata abount nutrients usually in a nutrients collection
type Nutrient struct {
	NutrientID uint   `json:"id" binding:"required"`
	Nutrientno int    `json:"nutrientno" binding:"required"`
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
	ID           string      `json:"_id" binding:"required"`
	Rev          string      `json:"_rev,omitempty"`
	FdcID        string      `json:"fdcId" binding:"required"`
	Upc          string      `json:"upc,omitempty"`
	Description  string      `json:"foodDescription" binding:"required"`
	Manufacturer string      `json:"company,omitempty"`
	Category     string      `json:"category,omitempty"`
	Source       string      `json:"Datasource"`
	Type         string      `json:"type"`
	Value        float64     `json:"valuePer100UnitServing"`
	Portion      string      `json:"portion,omitempty"`
	PortionValue float64     `json:"portionValue,omitempty"`
	Unit         string      `json:"unit"  binding:"required"`
	Derivation   *Derivation `json:"derivation,omitempty"`
	Nutrientno   int         `json:"nutrientNumber"`
	Nutrient     string      `json:"nutrientName"`
	Datapoints   int         `json:"datapoints,omitempty"`
	Min          float32     `json:"min,omitempty"`
	Max          float32     `json:"max,omitempty"`
}

// FoodGroup is the dictionary of FNDDS and SR food groups
type FoodGroup struct {
	ID          int32  `json:"id" binding:"required"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description" binding:"required"`
	LastUpdate  string `json:"lastUpdate,omitempty"`
	Type        string `json:"type" binding:"required"`
}
