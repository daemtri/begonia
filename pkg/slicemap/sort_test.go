package slicemap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	arr := []int{7, 102, 84, 12}
	Sort(arr, func(left, right int) bool {
		return left < right
	})
	assert.Equal(t, []int{7, 12, 84, 102}, arr)
}
