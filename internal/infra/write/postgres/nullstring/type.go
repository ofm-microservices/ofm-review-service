package nullstring

import "database/sql"

// Converter turns nullable SQL strings into plain Go strings.
type Converter interface {
	Convert(v sql.NullString) string
}
