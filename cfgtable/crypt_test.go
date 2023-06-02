package cfgtable

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func Test_Crypt(t *testing.T) {
	key := GenerateAesKey([]byte("xxx"))
	t.Log("key", string(key))
	data, err := AesCtrEncrypt([]byte("abc"), key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("加密后", data)
	time.Sleep(time.Second)
	origin, err := AesCtrEncrypt(data, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("解密后", string(origin))
	assert.Equal(t, string(origin), "abc")
}
