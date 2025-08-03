package nulltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
)

// NullInt64 represents a Cloud Spanner INT64 that may be NULL.
type NullInt64 struct {
	Int64 int64 // Int64 contains the value when it is non-NULL, and zero when NULL.
	Valid bool  // Valid is true if Int64 is not NULL.
}

// IsNull implements NullableValue.IsNull for NullInt64.
func (n NullInt64) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullInt64
func (n NullInt64) String() string {
	if !n.Valid {
		return nullString
	}
	return fmt.Sprintf("%v", n.Int64)
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullInt64.
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%v", n.Int64)), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullInt64.
func (n *NullInt64) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Int64 = int64(0)
		n.Valid = false
		return nil
	}
	num, err := strconv.ParseInt(string(payload), 10, 64)
	if err != nil {
		return fmt.Errorf("payload cannot be converted to int64: got %v", string(payload))
	}
	n.Int64 = num
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullInt64) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Int64, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullInt64) Scan(value interface{}) error {
	if value == nil {
		n.Int64, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullInt64: %v", p))
	case *int64:
		n.Int64 = *p
	case int64:
		n.Int64 = p
	case *NullInt64:
		n.Int64 = p.Int64
		n.Valid = p.Valid
	case NullInt64:
		n.Int64 = p.Int64
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullInt64) GormDataType() string {
	return "INT64"
}
