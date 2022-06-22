package gen_test

import (
	"github.com/stretchr/testify/require"
	"gpoker/gen"
	"testing"
)

func TestRandLowercaseString(t *testing.T) {
	// just check the length is equal to expected
	const expectedLen = 10
	for i := 0; i < expectedLen; i++ { // to reason to have expectedLen as a limit, using it just because
		require.Equal(t, expectedLen, len(gen.RandLowercaseString()))
	}
}
