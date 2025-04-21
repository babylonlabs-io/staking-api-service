package pkg

import (
	"strconv"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var input []int
		actual := Map(input, strconv.Itoa)
		assert.Nil(t, actual)
	})
	t.Run("slice with data", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6}
		expected := []string{"1", "2", "3", "4", "5", "6"}

		strings := Map(input, strconv.Itoa)
		assert.Equal(t, expected, strings)
	})
}
