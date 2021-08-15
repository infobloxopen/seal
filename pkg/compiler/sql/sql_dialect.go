package sqlcompiler

// SQLDialectEnum enumerates SQL dialects
type SQLDialectEnum int

// SQL dialects
const (
	DialectUnknown SQLDialectEnum = iota
	DialectPostgres
)

var dialectNames = []string{
	"DialectUnknown",
	"DialectPostgres",
}

// String satisfies fmt.Stringer interface
func (dia SQLDialectEnum) String() string {
	dint := int(dia)
	if dint < 0 || dint >= len(dialectNames) {
		dint = 0
	}
	return dialectNames[dint]
}
