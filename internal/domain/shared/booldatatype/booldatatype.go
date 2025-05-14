package booldatatype

import (
	"database/sql"
	"encoding/json"
)

// BoolDataType adalah tipe yang digunakan untuk menangani nilai 'Y' atau 'N' dari database
type BoolDataType struct {
	sql.NullString
}

// NewBoolDataType membuat BoolDataType baru dari string
func NewBoolDataType(s string) BoolDataType {
	return BoolDataType{
		NullString: sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

// MarshalJSON mengonversi nilai BoolDataType menjadi JSON boolean
func (b BoolDataType) MarshalJSON() ([]byte, error) {
	if !b.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(b.String == "Y")
}

// UnmarshalJSON mengonversi JSON boolean menjadi nilai BoolDataType
func (b *BoolDataType) UnmarshalJSON(data []byte) error {
	var result bool
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	if result {
		b.String = "Y"
	} else {
		b.String = "N"
	}
	b.Valid = true
	return nil
}

// SetNullIfEmpty mengatur nilai menjadi NULL jika kosong
func (b *BoolDataType) SetNullIfEmpty() {
	if b.String == "" {
		b.Valid = false
	}
}

// ToBool mengonversi BoolDataType ke nilai boolean
func (b *BoolDataType) ToBool() bool {
	return b.Valid && b.String == "Y"
}

// FromBool mengonversi nilai boolean ke BoolDataType
func FromBool(value bool) BoolDataType {
	if value {
		return BoolDataType{NullString: sql.NullString{String: "Y", Valid: true}}
	}
	return BoolDataType{NullString: sql.NullString{String: "N", Valid: true}}
}