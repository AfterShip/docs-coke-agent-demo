package nulltype

import (
	"bytes"
	"cloud.google.com/go/civil"
	"database/sql/driver"
	"errors"
	"fmt"
)

// NullDate represents a Cloud Spanner DATE that may be null.
type NullDate struct {
	Date  civil.Date // Date contains the value when it is non-NULL, and a zero civil.Date when NULL.
	Valid bool       // Valid is true if Date is not NULL.
}

// IsNull implements NullableValue.IsNull for NullDate.
func (n NullDate) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullDate
func (n NullDate) String() string {
	if !n.Valid {
		return nullString
	}
	return n.Date.String()
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullDate.
func (n NullDate) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%q", n.String())), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullDate.
func (n *NullDate) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Date = civil.Date{}
		n.Valid = false
		return nil
	}
	payload, err := trimDoubleQuotes(payload)
	if err != nil {
		return err
	}
	s := string(payload)
	t, err := civil.ParseDate(s)
	if err != nil {
		return fmt.Errorf("payload cannot be converted to civil.Date: got %v", string(payload))
	}
	n.Date = t
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullDate) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Date, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullDate) Scan(value interface{}) error {
	if value == nil {
		n.Date, n.Valid = civil.Date{}, false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullDate: %v", p))
	case *civil.Date:
		n.Date = *p
	case civil.Date:
		n.Date = p
	case *NullDate:
		n.Date = p.Date
		n.Valid = p.Valid
	case NullDate:
		n.Date = p.Date
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullDate) GormDataType() string {
	return "DATE"
}

func trimDoubleQuotes(payload []byte) ([]byte, error) {
	if len(payload) <= 1 || payload[0] != '"' || payload[len(payload)-1] != '"' {
		return nil, fmt.Errorf("payload is too short or not wrapped with double quotes: got %q", string(payload))
	}
	// Remove the double quotes at the beginning and the end.
	return payload[1 : len(payload)-1], nil
}
