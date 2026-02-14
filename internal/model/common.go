package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONBArray represents a JSONB array type for PostgreSQL
type JSONBArray []string

// Scan implements sql.Scanner for JSONBArray
func (j *JSONBArray) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONBArray: value is not []byte")
	}

	var arr []string
	if err := json.Unmarshal(bytes, &arr); err != nil {
		return err
	}

	*j = arr
	return nil
}

// Value implements driver.Valuer for JSONBArray
func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}
