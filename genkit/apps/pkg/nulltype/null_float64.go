package nulltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
)

// NullFloat64 represents a Cloud Spanner FLOAT64 that may be NULL.
type NullFloat64 struct {
	Float64 float64 // Float64 contains the value when it is non-NULL, and zero when NULL.
	Valid   bool    // Valid is true if Float64 is not NULL.
}

// IsNull implements NullableValue.IsNull for NullFloat64.
func (n NullFloat64) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullFloat64
func (n NullFloat64) String() string {
	if !n.Valid {
		return nullString
	}
	return fmt.Sprintf("%v", n.Float64)
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullFloat64.
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%v", n.Float64)), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullFloat64.
func (n *NullFloat64) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Float64 = float64(0)
		n.Valid = false
		return nil
	}
	num, err := strconv.ParseFloat(string(payload), 64)
	if err != nil {
		return fmt.Errorf("payload cannot be converted to float64: got %v", string(payload))
	}
	n.Float64 = num
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullFloat64) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Float64, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullFloat64) Scan(value interface{}) error {
	if value == nil {
		n.Float64, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullFloat64: %v", p))
	case *float64:
		n.Float64 = *p
	case float64:
		n.Float64 = p
	case *NullFloat64:
		n.Float64 = p.Float64
		n.Valid = p.Valid
	case NullFloat64:
		n.Float64 = p.Float64
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullFloat64) GormDataType() string {
	return "FLOAT64"
}
