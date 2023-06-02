package cfgtable

import "encoding/json"

var _ ConfigInterface = &String[int]{}

type String[T any] struct {
	Config[T]
}

func (c *String[T]) Fetch() ([]byte, error) {
	return []byte(c.URL), nil
}

func (c *String[T]) Parse(r []byte) error {
	return json.Unmarshal(r, &c.JSON)
}
