// Package fdc describes food products data model
package fdc

// DocType provides a list of document types
type DocType int

// FULL etc define values for Foods formats
const (
	FULL      = "full"
	META      = "meta"
	SERVING   = "servings"
	NUTRIENTS = "nutrients"
)

// PHRASE etc defines values for Search Types
const (
	PHRASE   = "PHRASE"
	WILDCARD = "WILDCARD"
	REGEX    = "REGEX"
)

// SR is standard reference
const (
	SR DocType = iota
	FGSR
	FNDDS
	FGFNDDS
	FGGPC
	BFPD
	UNIT
	NUT
	DERV
	FOOD
	USER
	NUTDATA
)

//ToDocType -- convert a string to a DocType
func (dt *DocType) ToDocType(t string) DocType {
	switch t {
	case "SR":
		return SR
	case "FGSR":
		return FGSR
	case "FGFNDDS":
		return FGFNDDS
	case "FGGPC":
		return FGGPC
	case "FNDDS":
		return FNDDS
	case "BFPD":
		return BFPD
	case "UNIT":
		return UNIT
	case "NUT":
		return NUT
	case "NUTDATA":
		return NUTDATA
	case "DERV":
		return DERV
	case "FOOD":
		return FOOD
	case "USER":
		return USER
	default:
		return 999
	}
}

//ToString -- convert a DocType to a string
func (dt *DocType) ToString(t DocType) string {
	switch t {
	case SR:
		return "SR"
	case FGSR:
		return "FGSR"
	case FGFNDDS:
		return "FGFNDDS"
	case FNDDS:
		return "FNDDS"
	case FGGPC:
		return "FGGPC"
	case BFPD:
		return "BFPD"
	case UNIT:
		return "UNIT"
	case NUT:
		return "NUT"
	case NUTDATA:
		return "NUTDATA"
	case DERV:
		return "DERV"
	case FOOD:
		return "FOOD"
	case USER:
		return "USER"
	default:
		return ""
	}
}
