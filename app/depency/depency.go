package depency

import (
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

var (
	config *Config = &Config{
		Allows: map[string]map[string]mapset.Set[string]{},
	}
)

func SetConfig(c *Config) {
	config = c
}

func SetModuleConfig(module string, rules []string) {
	if _, ok := config.Allows[module]; !ok {
		config.Allows[module] = map[string]mapset.Set[string]{}
	}

	for _, rule := range rules {
		r := strings.SplitN(rule, ":", 2)
		kind := r[0]
		name := ""
		if len(r) == 2 {
			name = r[1]
		}
		if _, ok := config.Allows[module][kind]; !ok {
			config.Allows[module][kind] = mapset.NewSet[string]()
		}

		config.Allows[module][kind].Add(name)
	}
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
