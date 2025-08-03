package uuid

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/gofrs/uuid"
)

const (
	RegionUS = iota + 1
	RegionEU
	RegionGlobal
)

var ErrInvalidUUIDFormat = errors.New("invalid uuid format")

type UUIDV4 struct {
	uuid.UUID
}

// SetRegion would set region bits into uuid bytes
// UUID V4's 11-12 bytes were used for node id, it's ok that we changed the
// high 4 bits to the fixed region. So the max number of regions was 15,
// but it's fine that we can extend if it's not enough.
// For more detail can see RFC4122: https://datatracker.ietf.org/doc/html/rfc4122
func (uuidV4 *UUIDV4) SetRegion(region byte) {
	uuidV4.UUID[12] = (uuidV4.UUID[12] & 0x0f) | (region << 4)
}

// Region retrieve the region byte from uuid v4 string
func (uuidV4 *UUIDV4) Region() byte {
	// We use the variant = FUTURE as the special flag which
	// brings the region flag, so we need to check the variant
	// before checking the region. Or it may get the wrong region
	// flag since the 12th byte was random in RFC4122 uuid v4.
	if uuidV4.Variant() == uuid.VariantFuture {
		return uuidV4.UUID[12] >> 4
	}
	return RegionUS
}

// NewV4FromString would create uuid v4 object from string
func NewV4FromString(uuid string) (*UUIDV4, error) {
	return NewV4FromBytes([]byte(uuid))
}

// NewV4FromBytes would create uuid v4 object from bytes
func NewV4FromBytes(uuidBytes []byte) (*UUIDV4, error) {
	uuidV4 := &UUIDV4{}
	if len(uuidBytes) != 32 && len(uuidBytes) != 36 {
		return nil, ErrInvalidUUIDFormat
	}

	var err0, err1, err2, err3, err4 error
	if len(uuidBytes) == 32 {
		_, err0 = hex.Decode(uuidV4.UUID[0:4], uuidBytes[0:8])
		_, err1 = hex.Decode(uuidV4.UUID[4:6], uuidBytes[8:12])
		_, err2 = hex.Decode(uuidV4.UUID[6:8], uuidBytes[12:16])
		_, err3 = hex.Decode(uuidV4.UUID[8:10], uuidBytes[16:20])
		_, err4 = hex.Decode(uuidV4.UUID[10:], uuidBytes[20:])
	} else {
		// Compatible with raw uuidBytes v4 format
		if uuidBytes[8] != '-' || uuidBytes[13] != '-' || uuidBytes[18] != '-' || uuidBytes[23] != '-' {
			return nil, ErrInvalidUUIDFormat
		}
		_, err0 = hex.Decode(uuidV4.UUID[0:4], uuidBytes[0:8])
		_, err1 = hex.Decode(uuidV4.UUID[4:6], uuidBytes[9:13])
		_, err2 = hex.Decode(uuidV4.UUID[6:8], uuidBytes[14:18])
		_, err3 = hex.Decode(uuidV4.UUID[8:10], uuidBytes[19:23])
		_, err4 = hex.Decode(uuidV4.UUID[10:], uuidBytes[24:])
	}
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, errors.New("decode uuid error")
	}
	if uuidV4.Version() != uuid.V4 {
		return nil, errors.New("uuid version must be v4")
	}
	return uuidV4, nil
}

// GenerateUUIDV4 would generate the uuid with the us region flag
func GenerateUUIDV4() string {
	uuidV4, _ := GenerateRegionUUIDV4(RegionUS)
	return uuidV4
}

// GenerateRegionUUIDV4 would generate the uuid with region flags
func GenerateRegionUUIDV4(region byte) (string, error) {
	if region != RegionUS && region != RegionEU && region != RegionGlobal {
		return "", errors.New("invalid region")
	}

	uuidV4 := UUIDV4{UUID: uuid.Must(uuid.NewV4())}
	// We need to change the variant to future since the region byte was random
	// bits in RFC4122 v4.
	uuidV4.SetVariant(uuid.VariantFuture)
	uuidV4.SetRegion(region)
	return strings.Replace(uuidV4.String(), "-", "", -1), nil
}
