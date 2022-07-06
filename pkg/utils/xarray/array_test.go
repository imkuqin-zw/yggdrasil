package xarray

import (
	"fmt"
	"testing"
)

func TestShuffleStringArray(t *testing.T) {
	fmt.Println(ShuffleStringArray([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9}))
	fmt.Println(ShuffleStringArray([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9}))
	fmt.Println(ShuffleStringArray([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9}))
	fmt.Println(ShuffleStringArray([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9}))
	fmt.Println(ShuffleStringArray([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9}))
}
