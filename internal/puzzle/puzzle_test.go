package puzzle_test

import (
	"15-puzzle/internal/puzzle"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSolvable(t *testing.T) {
	table := []struct {
		puzzle   [16]byte
		solvable bool
	}{
		{[16]byte{
			13, 2, 10, 3,
			1, 12, 8, 4,
			5, 0, 9, 6,
			15, 14, 11, 7},
			true},
		{[16]byte{
			6, 13, 7, 10,
			8, 9, 11, 0,
			15, 2, 12, 5,
			14, 3, 1, 4},
			true},
		{[16]byte{
			12, 1, 10, 2,
			7, 11, 4, 14,
			5, 0, 9, 15,
			8, 13, 6, 3},
			true},
		{[16]byte{ // not solvable permutation
			3, 9, 1, 15,
			14, 11, 4, 6,
			13, 0, 10, 12,
			2, 7, 8, 5},
			false},
		{[16]byte{ // same as before but tiles 1 and 2 are switched
			3, 9, 2, 15,
			14, 11, 4, 6,
			13, 0, 10, 12,
			1, 7, 8, 5},
			true},
	}
	for i := range table {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			solvable, err := puzzle.IsSolvable(table[i].puzzle)
			assert.NoError(t, err)
			assert.Equal(t, table[i].solvable, solvable)
		})
	}
	_, err := puzzle.IsSolvable([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 11, 13, 14, 15, 16})
	assert.Error(t, err)
	_, err = puzzle.IsSolvable([16]byte{0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0})
	assert.Error(t, err)
}
