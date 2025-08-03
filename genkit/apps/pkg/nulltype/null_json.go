package nulltype

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

var (
	jsonNullBytes = []byte("null")
)

// NullJSON represents a Cloud Spanner JSON that may be NULL.
//
// This type must always be used when encoding values to a JSON column in Cloud
// Spanner.
//
// NullJSON does not implement the driver.Valuer and sql.Scanner interfaces, as
// the underlying value can be anything. This means that the type NullJSON must
// also be used when calling sql.Row#Scan(dest ...interface{}) for a JSON
// column.
type NullJSON struct {
	Value interface{} // Val contains the value when it is non-NULL, and nil when NULL.
	Valid bool        // Valid is true if Json is not NULL.
}

// IsNull implements NullableValue.IsNull for NullJSON.
func (n NullJSON) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullJSON.
func (n NullJSON) String() string {
	if !n.Valid {
		return nullString
	}
	b, err := jsoniter.Marshal(n.Value)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return fmt.Sprintf("%v", string(b))
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullJSON.
func (n NullJSON) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return jsoniter.Marshal(n.Value)
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullJSON.
func (n *NullJSON) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Valid = false
		return nil
	}
	var v interface{}
	err := jsoniter.Unmarshal(payload, &v)
	if err != nil {
		return fmt.Errorf("payload cannot be converted to a struct: got %v, err: %w", string(payload), err)
	}
	n.Value = v
	n.Valid = true
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullJSON) GormDataType() string {
	return "JSON"
}
