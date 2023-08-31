package status

import (
	"errors"
	"fmt"
	"testing"

	errors2 "github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
)

func TestNew(t *testing.T) {
	st := New(code.Code_NOT_FOUND, errors.New("fdafds"))
	fmt.Printf("%+v\n", st)
	st2 := New(code.Code_NOT_FOUND, errors2.New("fdafds"))
	fmt.Printf("%+v\n", st2)
}
