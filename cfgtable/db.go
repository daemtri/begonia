package cfgtable

var _ ConfigInterface = &DB{}

type DB struct {
	// URL 配置远程下载地址
	Value string `flag:"db" usage:"配置值"`
}

func (c *DB) URI() string {
	return c.Value
}

func (c *DB) Fetch() ([]byte, error) {
	return []byte(c.Value), nil
}

func (c *DB) Parse(_ []byte) error {
	return nil
}
