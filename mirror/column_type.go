package mirror

import "reflect"

// ColumnType contains the name and type of a column.
type ColumnType struct {
	Name string

	HasNullable       bool
	HasLength         bool
	HasPrecisionScale bool

	Nullable     bool
	Length       int64
	DatabaseType string
	Precision    int64
	Scale        int64
	ScanType     reflect.Type
}
