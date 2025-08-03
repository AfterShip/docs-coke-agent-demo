package nulltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
)

// NullBool represents a Cloud Spanner BOOL that may be NULL.
type NullBool struct {
	Bool  bool // Bool contains the value when it is non-NULL, and false when NULL.
	Valid bool // Valid is true if Bool is not NULL.
}

// IsNull implements NullableValue.IsNull for NullBool.
func (n NullBool) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullBool
func (n NullBool) String() string {
	if !n.Valid {
		return nullString
	}
	return fmt.Sprintf("%v", n.Bool)
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullBool.
func (n NullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%v", n.Bool)), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullBool.
func (n *NullBool) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Bool = false
		n.Valid = false
		return nil
	}
	b, err := strconv.ParseBool(string(payload))
	if err != nil {
		return fmt.Errorf("payload cannot be converted to bool: got %v", string(payload))
	}
	n.Bool = b
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullBool) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Bool, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullBool) Scan(value interface{}) error {
	if value == nil {
		n.Bool, n.Valid = false, false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullBool: %v", p))
	case *bool:
		n.Bool = *p
	case bool:
		n.Bool = p
	case *NullBool:
		n.Bool = p.Bool
		n.Valid = p.Valid
	case NullBool:
		n.Bool = p.Bool
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullBool) GormDataType() string {
	return "BOOL"
}
