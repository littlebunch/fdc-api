// Package fdc describes food products data model
package fdc

import (
	"io/ioutil"
	"log"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Food reflects JSON used to transfer BFPD foods data from JIFSAN to NAL
type Food struct {
	UpdatedAt       time.Time      `json:"lastChangeDateTime"`
	FdcID           string         `json:"fdcId" binding:"required"`
	Upc             string         `json:"upc"`
	Description     string         `json:"foodDescription" binding:"required"`
	Source          string         `json:"dataSource"`
	PublicationDate time.Time      `json:"publicationDateTime"`
	Ingredients     string         `json:"ingredients"`
	Manufacturer    string         `json:"company"`
	Servings        []Serving      `json:"servingSizes"`
	Nutrients       []NutrientData `json:"nutrients"`
}

// Serving describes a list nutrients for a given state, weight and amount
type Serving struct {
	Nutrientbasis string  `json:"100UnitNutrientBasis"`
	Description   string  `json:"householdServingUom"`
	Servingstate  string  `json:"servingState"`
	Weight        float32 `json:"weightInGmOrMl"`
	Servingamount float32 `json:"householdServingValue"`
}

// Nutrient is metadata abount nutrients usually in a nutrients collection
type Nutrient struct {
	Nutrientno uint   `json:"nutrientno" binding:"required"`
	Tagname    string `json:"tagname"`
	Name       string `json:"name"  binding:"required"`
	Unit       string `json:"unit"  binding:"required"`
	Type       string `json:"type"  binding:"required"`
}

// Derivation is a code for describing how nutrient values are derived
type Derivation struct {
	Code string `json:"code"`
}

// NutrientData is the list of nutrient values attached to Serving
type NutrientData struct {
	Value      float32 `json:"valuePer100UnitServing"`
	Unit       string  `json:"unit"  binding:"required"`
	Derivation string  `json:"derivation"`
	Nutrientno uint    `json:"nutrientNumber"`
	Nutrient   string  `json:"nutrientName"`
}

// FoodGroup is the dictionary of FNDDS and SR food groups
type FoodGroup struct {
	Code        string `json:"code" binding:"required"`
	Description string `json:"description" binding:"required"`
	LastUpdate  string `json:"lastUpdate"`
	Type        string `json:"type" binding:"required"`
}

//Config provides basic MySQL, Elasticsearch and Nsq configuration properties for bfpd web and microservices.  Properties are normally read in from a YAML file or the environment
type Config struct {
	CouchDb CouchDb
}

// CouchDb configuration for connecting, reading and writing Couchbase nodes
type CouchDb struct {
	URL        string
	ReplicaSet string
	Bucket     string
	User       string
	Pwd        string
}

// Defaults sets values for Elastic and Nsq configuration properties if none have been provided.
func (cs *Config) Defaults() {
	if cs.CouchDb.URL == "" {
		cs.CouchDb.URL = "localhost"
	}
	if cs.CouchDb.Bucket == "" {
		cs.CouchDb.Bucket = "gnutdata"
	}
}

// GetConfig reads config from a file
func (cs *Config) GetConfig(c *string) {
	raw, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Println(err.Error())
	}
	if err = yaml.Unmarshal(raw, cs); err != nil {
		log.Println(err.Error())
	}
	cs.Defaults()
}
