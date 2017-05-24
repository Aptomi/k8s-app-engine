package slinga

import (
	"bytes"
	"log"
)

// ServiceUsageState contains detailed tracing information for specific allocations
type ServiceUsageTracing struct {
	buf    *bytes.Buffer
	logger *log.Logger
	trace  bool
}

func NewServiceUsageTracing() *ServiceUsageTracing {
	buf := new(bytes.Buffer)
	logger := log.New(buf, "", 0)
	result := &ServiceUsageTracing{
		buf:    buf,
		logger: logger}
	return result
}

func (tracing *ServiceUsageTracing) do(t bool) *ServiceUsageTracing {
	tracing.trace = t
	return tracing
}

func (tracing *ServiceUsageTracing) log(depth int, format string, args ...interface{}) {
	if tracing.trace {
		indent := ""
		for n := 0; n <= 4*depth; n++ {
			indent = indent + " "
		}
		format = indent + format
		tracing.logger.Printf(format, args...)
		log.Printf(format, args...)
	}
}

func (tracing *ServiceUsageTracing) newline() {
	tracing.log(0, "\n")
}
