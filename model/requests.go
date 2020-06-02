package fdc

// BrowseResult is returned from the browse endpoints
type BrowseResult struct {
	Count int32         `json:"count"`
	Start int32         `json:"start"`
	Max   int32         `json:"max"`
	Items []interface{} `json:"items"`
}

// BrowseNutrientReport is returned from the nutrients report endpoing
type BrowseNutrientReport struct {
	Request NutrientReportRequest `json:"request"`
	Items   []interface{}         `json:"foods"`
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

// NutrientReportRequest wraps a POST nutrient report
type NutrientReportRequest struct {
	Page      int     `json:"page"`
	Max       int     `json:"max"`
	Nutrient  int     `json:"nutrientno" binding:"required"`
	FoodGroup string  `json:"foodGroup,omitEmpty"`
	Sort      string  `json:"sort,omitEmpty"`
	Order     string  `json:"order,omitEmpty"`
	ValueGTE  float64 `json:"valueGTE"`
	ValueLTE  float64 `json:"valueLTE"`
}

// SearchRequest wraps a POST search
type SearchRequest struct {
	Query       string `json:"q" binding:"required"`
	SearchField string `json:"searchfield,omitEmpty"`
	Page        int    `json:"page"`
	Max         int    `json:"max"`
	Sort        string `json:"sort,omitEmpty"`
	SearchType  string `json:"searchtype,omitEmpty"`
	FoodGroup   string `json:"foodgroup,omitEmpty"`
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

// NutrientFoodBrowse is returned from the food nutrient endpoints
type NutrientFoodBrowse struct {
	FdcID        string                   `json:"fdcId" binding:"required"`
	Upc          string                   `json:"upc,omitempty"`
	Description  string                   `json:"foodDescription" binding:"required"`
	Manufacturer string                   `json:"company,omitempty"`
	Category     string                   `json:"category,omitempty"`
	Portion      string                   `json:"portion,omitempty"`
	Nutrients    []NutrientFoodBrowseItem `json:"nutrients"`
}

// NutrientFoodBrowseItem is the list of nutrient data returned by the food nutrient endpoints
type NutrientFoodBrowseItem struct {
	Value        float64     `json:"valuePer100UnitServing"`
	Unit         string      `json:"unit"  binding:"required"`
	Derivation   *Derivation `json:"derivation,omitempty"`
	Nutrientno   int         `json:"nutrientNumber"`
	Nutrient     string      `json:"nutrientName"`
	PortionValue float64     `json:"valuePerPortion"`
}

// NutrientReportData is an item returned in a nutrient report
type NutrientReportData struct {
	FdcID           string  `json:"fdcId" binding:"required"`
	Upc             string  `json:"upc"`
	FoodDescription string  `json:"foodDescription"`
	Value           float64 `json:"valuePer100UnitServing"`
	Portion         string  `json:"portion,omitempty"`
	PortionValue    float64 `json:"valuePerPortion"`
	Unit            string  `json:"unit"`
	Type            string  `json:"type"`
}
