package nullstring

import "database/sql"

type converter struct{}

// New constructs a nullable-string converter.
func New() Converter {
	return converter{}
}

func (converter) Convert(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}
