package est

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_fitDx_simple(t *testing.T) {
	testData := []int{
		34, 34, 34, 34, 34, 26, 0, 34, 1, 1, 0, 0, 20, 0, 34, 34, 34, 34, 25, 34, 34,
		34, 34, 34, 34, 34, 34, 34, 22, 0, 34, 34, 28, 34, 34, 26, 27, 34, 34, 34, 34,
		34, 34, 0, 0, 34, 34, 34, 0, 34, 34, 34, 34, 1, 34, 34, 22, 34, 34, 34, 34, 0,
		34, 0, 34, 34, 26, 34, 34, 34, 3, 34, 34, 32, 34, 34, 34, 7, 0, 34, 0, 34, 1,
		34, 34, 0, 34, 34, 5, 34, 5, 34, 27, 0, 0, 34, 34, 34, 34, 34, 32, 31, 34, 34,
		29, 25, 34, 10, 0, 6, 0, 34, 0, 34, 1, 24, 34, 34, 35,
	}

	res, err := fitDx(testData)
	require.NoError(t, err)
	assert.Equal(t, 119, len(res))
	assert.Equal(t, 34, res[10])
}

func Test_fitDx_negative(t *testing.T) {
	testData := []int{
		-9, -9, -9, -9, -9, -9, -9, -9, -9, -9,
		-10, -10, -10, -10, -10, -10, -10, -10, -10, -10,
	}

	res, err := fitDx(testData)
	require.NoError(t, err)
	assert.Equal(t, 20, len(res))
	assert.Equal(t, -9, res[5])
	assert.Equal(t, -10, res[len(res)-5])
}

func Test_fitDx_rounding(t *testing.T) {
	testData := []int{
		9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
		10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
	}

	res, err := fitDx(testData)
	require.NoError(t, err)
	assert.Equal(t,
		[]int{
			9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 10, 9, 10,
			9, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 11, 11,
		},
		res,
	)
}

func Test_fitDx_rounding_negative(t *testing.T) {
	testData := []int{
		-9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9,
		-10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10,
	}

	res, err := fitDx(testData)
	require.NoError(t, err)
	assert.Equal(t,
		[]int{
			-9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -10, -9,
			-10, -9, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -10, -11, -11,
		},
		res,
	)
}
