package nulltype

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"math/big"
)

const (
	// NumericScaleDigits is the maximum number of digits after the decimal
	// point in a NUMERIC value.
	NumericScaleDigits = 9
)

// NullNumeric represents a Cloud Spanner Numeric that may be NULL.
type NullNumeric struct {
	Numeric big.Rat // Numeric contains the value when it is non-NULL, and a zero big.Rat when NULL.
	Valid   bool    // Valid is true if Numeric is not NULL.
}

// IsNull implements NullableValue.IsNull for NullNumeric.
func (n NullNumeric) IsNull() bool {
	return !n.Valid
}

// String implements Stringer.String for NullNumeric
func (n NullNumeric) String() string {
	if !n.Valid {
		return nullString
	}
	return fmt.Sprintf("%v", NumericString(&n.Numeric))
}

// MarshalJSON implements json.Marshaler.MarshalJSON for NullNumeric.
func (n NullNumeric) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return []byte(fmt.Sprintf("%q", NumericString(&n.Numeric))), nil
	}
	return jsonNullBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON for NullNumeric.
func (n *NullNumeric) UnmarshalJSON(payload []byte) error {
	if payload == nil {
		return fmt.Errorf("payload should not be nil")
	}
	if bytes.Equal(payload, jsonNullBytes) {
		n.Numeric = big.Rat{}
		n.Valid = false
		return nil
	}
	payload, err := trimDoubleQuotes(payload)
	if err != nil {
		return err
	}
	s := string(payload)
	val, ok := (&big.Rat{}).SetString(s)
	if !ok {
		return fmt.Errorf("payload cannot be converted to big.Rat: got %v", string(payload))
	}
	n.Numeric = *val
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullNumeric) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Numeric, nil
}

// Scan implements the sql.Scanner interface.
func (n *NullNumeric) Scan(value interface{}) error {
	if value == nil {
		n.Numeric, n.Valid = big.Rat{}, false
		return nil
	}
	n.Valid = true
	switch p := value.(type) {
	default:
		return errors.New(fmt.Sprintf("invalid type for NullNumeric: %v", p))
	case *big.Rat:
		n.Numeric = *p
	case big.Rat:
		n.Numeric = p
	case *NullNumeric:
		n.Numeric = p.Numeric
		n.Valid = p.Valid
	case NullNumeric:
		n.Numeric = p.Numeric
		n.Valid = p.Valid
	}
	return nil
}

// GormDataType is used by gorm to determine the default data type for fields with this type.
func (n NullNumeric) GormDataType() string {
	return "NUMERIC"
}

// NumericString returns a string representing a *big.Rat in a format compatible
// with Spanner SQL. It returns a floating-point literal with 9 digits after the
// decimal point.
func NumericString(r *big.Rat) string {
	return r.FloatString(NumericScaleDigits)
}
