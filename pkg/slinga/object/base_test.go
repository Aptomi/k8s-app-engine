package object

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKey(t *testing.T) {
	correctKey := KeyFromParts("72b062c1-7fcf-11e7-ab09-acde48001122", 42)

	assert.Equal(t, UID("72b062c1-7fcf-11e7-ab09-acde48001122"), correctKey.GetUID(), "Correct UID expected")
	assert.Equal(t, Generation(42), correctKey.GetGeneration(), "Correct Generation expected")

	noGenerationKey := Key("72b062c1-7fcf-11e7-ab09-acde48001122")

	assert.Panics(t, func() { noGenerationKey.GetUID() }, "Panic expected if key is incorrect")
	assert.Panics(t, func() { noGenerationKey.GetGeneration() }, "Panic expected if key is incorrect")

	invalidGenerationKey := Key("72b062c1-7fcf-11e7-ab09-acde48001122" + KeySeparator + "bad")

	assert.Equal(t, UID("72b062c1-7fcf-11e7-ab09-acde48001122"), correctKey.GetUID(), "Correct UID expected")
	assert.Panics(t, func() { invalidGenerationKey.GetGeneration() }, "Panic expected if key is incorrect")
}

func loopAndCheckNewUUIDs(tb testing.TB, n int) {
	var prev UID
	for i := 0; i < n; i++ {
		next := NewUUID()
		if next == prev {
			tb.Fatal("UIDs should be different")
		}
		prev = next
	}
}

func TestNewUUID(t *testing.T) {
	loopAndCheckNewUUIDs(t, 100000)
}

func BenchmarkNewUUID(b *testing.B) {
	loopAndCheckNewUUIDs(b, b.N)
}
