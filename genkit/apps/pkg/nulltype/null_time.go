package nulltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

// NullTime represents a Cloud Spanner TIMESTAMP that may be null.
type NullTime struct {
	Time  time.Time // Time contains the value when it is non-NULL, and a zero time.Time when NULL.
	Valid bool      // Valid is true if Time is not NULL.
}

// IsNull implements NullableValue.IsNull for NullTime.
func (n NullTime) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullTime
func (n NullTime) String() string {
	if !n.Valid {
		return nullString
	}
	return n.Time.Format(time.RFC3339Nano)
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullTime.
func (n NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%q", n.String())), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullTime.
func (n *NullTime) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Time = time.Time{}
		n.Valid = false
		return nil
	}
	payload, err := trimDoubleQuotes(payload)
	if err != nil {
		return err
	}
	s := string(payload)
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return fmt.Errorf("payload cannot be converted to time.Time: got %v", string(payload))
	}
	n.Time = t
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Time, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullTime) Scan(value interface{}) error {
	if value == nil {
		n.Time, n.Valid = time.Time{}, false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullTime: %v", p))
	case *time.Time:
		n.Time = *p
	case time.Time:
		n.Time = p
	case *NullTime:
		n.Time = p.Time
		n.Valid = p.Valid
	case NullTime:
		n.Time = p.Time
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullTime) GormDataType() string {
	return "TIMESTAMP"
}
