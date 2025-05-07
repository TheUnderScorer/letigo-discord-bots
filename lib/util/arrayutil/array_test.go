package arrayutil_test

import (
	"github.com/stretchr/testify/assert"
	"lib/util/arrayutil"
	"testing"
)

func TestReverseSlice(t *testing.T) {
	t.Run("reverse slice", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		expectedResult := []int{5, 4, 3, 2, 1}
		result := arrayutil.ReverseSlice(slice)

		assert.Equal(t, expectedResult, result)
	})
}
