package nulltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

const (
	nullString = ""
)

// NullString represents a Cloud Spanner STRING that may be NULL.
type NullString struct {
	StringVal string // StringVal contains the value when it is non-NULL, and an empty string when NULL.
	Valid     bool   // Valid is true if StringVal is not NULL.
}

// IsNull implements NullableValue.IsNull for NullString.
func (n NullString) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullString
func (n NullString) String() string {
	if !n.Valid {
		return nullString
	}
	return n.StringVal
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullString.
func (n NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%q", n.StringVal)), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullString.
func (n *NullString) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.StringVal = ""
		n.Valid = false
		return nil
	}
	var s *string
	if err := jsoniter.Unmarshal(payload, &s); err != nil {
		return err
	}
	if s != nil {
		n.StringVal = *s
		n.Valid = true
	} else {
		n.StringVal = ""
		n.Valid = false
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullString) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.StringVal, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullString) Scan(value interface{}) error {
	if value == nil {
		n.StringVal, n.Valid = "", false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullString: %v", p))
	case *string:
		n.StringVal = *p
	case string:
		n.StringVal = p
	case *NullString:
		n.StringVal = p.StringVal
		n.Valid = p.Valid
	case NullString:
		n.StringVal = p.StringVal
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullString) GormDataType() string {
	return "STRING(MAX)"
}
