package est

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test -count 1 -v -race -timeout 30s -run Test_Ransac github.com/jo-m/trainbot/pkg/est

func Test_cleanupDx(t *testing.T) {
	testData := []int{
		34, 34, 34, 34, 34, 26, 0, 34, 1, 1, 0, 0, 20, 0, 34, 34, 34, 34, 25, 34, 34,
		34, 34, 34, 34, 34, 34, 34, 22, 0, 34, 34, 28, 34, 34, 26, 27, 34, 34, 34, 34,
		34, 34, 0, 0, 34, 34, 34, 0, 34, 34, 34, 34, 1, 34, 34, 22, 34, 34, 34, 34, 0,
		34, 0, 34, 34, 26, 34, 34, 34, 3, 34, 34, 32, 34, 34, 34, 7, 0, 34, 0, 34, 1,
		34, 34, 0, 34, 34, 5, 34, 5, 34, 27, 0, 0, 34, 34, 34, 34, 34, 32, 31, 34, 34,
		29, 25, 34, 10, 0, 6, 0, 34, 0, 34, 1, 24, 34, 34, 35, 0, 0, 0, 0, 0, 0, 0,
	}

	clean, err := cleanupDx(testData)
	require.NoError(t, err)
	fmt.Println(clean)
	// TODO: check results
}
