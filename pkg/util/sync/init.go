package sync

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Init is a helper to perform initializer function only single time and return saved error if any
type Init struct {
	m    sync.Mutex
	done uint32
	err  error
}

// Do calls function initializer only for the first time for this instance of Init and saves error returned from it
// if any. Do returns nil if no error happened during the initializer execution or it returns error. Even in case of
// error during initializer execution it will not be called second time, error will be saved and return each time Do
// called. If initializer panics, it'll be recovered and captured as error and returned by future Do calls.
func (init *Init) Do(initializer func() error) (err error) {
	// if init already done just returned stored error
	if atomic.LoadUint32(&init.done) == 1 {
		return init.err
	}

	// Slow path if init isn't done yet
	init.m.Lock()
	defer init.m.Unlock()

	// Check if init is still not done
	if init.done == 0 {
		// We consider init done if entered this section
		defer atomic.StoreUint32(&init.done, 1)

		// Convert panic into error and store it as init error
		defer func() {
			if recoveredErr := recover(); recoveredErr != nil {
				// save panic as error to be returned by next Init.Do call
				init.err = fmt.Errorf("panic during init: %s", recoveredErr)

				// override error returned from Init.Do
				err = init.err
			}
		}()

		// run provided init function and store error returned from it to be returned by next Init.Do call
		init.err = initializer()
		if init.err != nil {
			init.err = fmt.Errorf("error during init: %s", init.err)
		}
	}

	return init.err
}
