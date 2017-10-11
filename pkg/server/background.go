package server

import "fmt"

type job struct {
	name     string
	errors   chan string
	f        func()
	infinite bool
}

func (j *job) complete() {
	if r := recover(); r != nil {
		j.errors <- fmt.Sprintf("Background job '%s' failed with error (panic): %s", j.name, r)
	} else if j.infinite {
		j.errors <- fmt.Sprintf("Infinite background job '%s' completed without error", j.name)
	}
}

func (j *job) start() {
	defer j.complete()
	j.f()
}

func (s *Server) runInBackground(name string, infinite bool, f func()) {
	p := job{name, s.backgroundErrors, f, infinite}
	go p.start()
}

func (s *Server) wait() {
	err := <-s.backgroundErrors
	panic(err)
}
