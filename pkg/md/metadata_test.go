package md

import (
	"fmt"
	"testing"
)

func Test_mpCap(t *testing.T) {
	md := make(map[string][]string)
	for i := 0; i < 1000; i++ {
		changeMap(md, i)
		fmt.Println(md)
	}

}

func changeMap(md map[string][]string, i int) {
	md[fmt.Sprintf("%d", i)] = []string{fmt.Sprintf("%d", i)}
}
