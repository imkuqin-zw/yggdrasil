package errgroup

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testFunc(context.Context) error {
	time.Sleep(time.Second)
	return nil
}

func TestGroup_GOMAXPROCS(t *testing.T) {
	g1 := Group{}
	now := time.Now()
	g1.Go(testFunc)
	g1.Go(testFunc)
	g1.Go(testFunc)
	g1.Go(testFunc)
	_ = g1.Wait()
	sec := math.Round(time.Since(now).Seconds())
	require.Equal(t, float64(1), sec, "not limit go process")

	g2 := Group{}
	g2.GOMAXPROCS(2)
	now = time.Now()
	g2.Go(testFunc)
	g2.Go(testFunc)
	g2.Go(testFunc)
	g2.Go(testFunc)
	_ = g2.Wait()
	sec = math.Round(time.Since(now).Seconds())
	require.Equal(t, float64(2), sec, "limit go process")
}

func TestGroup_Normal(t *testing.T) {
	var (
		num = make([]int, 2)
		g   Group
		err error
	)
	for i := 0; i < 2; i++ {
		num[i] = i
	}
	g.Go(func(context.Context) (err error) {
		num[0]++
		return
	})
	g.Go(func(context.Context) (err error) {
		num[1]++
		return
	})
	err = g.Wait()
	require.Equal(t, nil, err, "wait result")
	require.Equal(t, 1, num[0], "num 0")
	require.Equal(t, 2, num[1], "num 1")

}

func TestGroup_Error(t *testing.T) {
	err1 := errors.New("errgroup_test: 1")
	err2 := errors.New("errgroup_test: 2")

	cases := []struct {
		name string
		errs []error
	}{
		{errs: []error{}},
		{errs: []error{nil}},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var g Group

			var firstErr error
			for _, err := range tc.errs {
				err := err
				g.Go(func(context.Context) error { return err })

				if firstErr == nil && err != nil {
					firstErr = err
				}
				gErr := g.Wait()
				require.Equal(t, firstErr, gErr)
			}
		})
	}
}

func TestGroup_Cancel(t *testing.T) {
	g := WithCancel(context.Background())
	g.Go(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("boom")
	})
	var doneErr error
	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		doneErr = ctx.Err()
		return doneErr
	})
	_ = g.Wait()
	require.Equal(t, context.Canceled, doneErr)
}
