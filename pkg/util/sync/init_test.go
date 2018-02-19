package sync

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	testInitIterations = 1000
)

//tests := []struct {
//	name string
//}{
//	// TODO: test cases
//}
//for _, test := range tests {
//	t.Run(test.name, func(t *testing.T) {
//
//	})
//}

func TestInit(t *testing.T) {
	var count uint32
	init := &Init{}

	testInitDo(t, &count, init, func() error {
		return nil
	}, nil)

	testInitDo(t, &count, init, func() error {
		return errors.New("Init.Do called twice")
	}, nil)

	testInitDo(t, &count, init, func() error {
		panic("Init.Do called twice")
	}, nil)
}

func TestInitError(t *testing.T) {
	var count uint32
	init := &Init{}

	testInitDo(t, &count, init, func() error {
		return errors.New("some error")
	}, errors.New("error during init: some error"))
}

func TestInitPanic(t *testing.T) {
	var count uint32
	init := &Init{}

	testInitDo(t, &count, init, func() error {
		panic("some panic")
	}, errors.New("panic during init: some panic"))
}

func testInitDo(t *testing.T, count *uint32, init *Init, initializer func() error, expected error) {
	t.Helper()

	var wg sync.WaitGroup
	wg.Add(testInitIterations)

	results := make(chan error, testInitIterations)

	for i := 0; i < testInitIterations; i++ {
		go func() {
			defer wg.Done()

			results <- init.Do(func() error {
				atomic.AddUint32(count, 1)

				return initializer()
			})
		}()
	}

	wg.Wait()
	close(results)

	assert.Equal(t, uint32(1), *count, "Initializer function should be called only once")
	assert.Len(t, results, testInitIterations, "Number of captured results should be equal to number of iterations")
	for initErr := range results {
		assert.Equal(t, expected, initErr, "Error returned by Init.Do should be the same as returned by initializer function")
	}
}
