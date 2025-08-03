package nulltype

// NullableValue is the interface implemented by all null value wrapper types.
type NullableValue interface {
	// IsNull returns true if the underlying database value is null.
	IsNull() bool
}
