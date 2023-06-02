package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type Json[T any] struct {
	V T
}

// Value 实现 driver.Valuer
func (s Json[T]) Value() (driver.Value, error) {
	b, err := json.Marshal(&s.V)
	return string(b), err
}

// Scan 实现sql.Scanner
func (s *Json[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to scan Array value:", value))
	}
	if len(bytes) > 0 {
		return json.Unmarshal(bytes, &s.V)
	}
	*s = Json[T]{}
	return nil
}

// MarshalJSON json序列化接口
func (s *Json[T]) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(s.V)
	return b, err
}
