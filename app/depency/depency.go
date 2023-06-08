package depency

import (
	mapset "github.com/deckarep/golang-set/v2"
)

var (
	config *Config
)

func SetConfig(c *Config) {
	config = c
}

type Config struct {
	// Allows is a list of rules that allow access to a resource.
	//  module:kind:[names...]
	Allows map[string]map[string]mapset.Set[string] `json:"allows"`
}

func (c *Config) Allow(module, kind, name string) bool {
	if _, ok := c.Allows[module]; !ok {
		return false
	}

	if _, ok := c.Allows[module][kind]; !ok {
		return false
	}

	return c.Allows[module][kind].Contains(name)
}

func Allow(module, kind, name string) bool {
	return config.Allow(module, kind, name)
}
