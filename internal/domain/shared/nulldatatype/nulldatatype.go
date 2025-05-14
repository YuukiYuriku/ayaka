	package nulldatatype

import (
	"database/sql"
	"encoding/json"
)

type NullDataType struct {
	sql.NullString
}

func NewNullStringDataType(s string) NullDataType {
	return NullDataType{
		NullString: sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

func (ns NullDataType) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullDataType) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == nil {
		ns.Valid = false
		return nil
	}
	ns.String = *s
	ns.Valid = true
	return nil
}

func (ns *NullDataType) SetNullIfEmpty() {
	if ns.String == "" {
		ns.Valid = false
	}
}