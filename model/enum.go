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

// SR is standard reference
const (
	SR DocType = iota
	FGSR
	FNDDS
	FGFNDDS
	BFPD
	UNIT
	NUT
	DERV
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
	case "FNDDS":
		return FNDDS
	case "BFPD":
		return BFPD
	case "UNIT":
		return UNIT
	case "NUT":
		return NUT
	case "DERV":
		return DERV
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
	case BFPD:
		return "BFPD"
	case UNIT:
		return "UNIT"
	case NUT:
		return "NUT"
	case DERV:
		return "DERV"
	default:
		return ""
	}
}
