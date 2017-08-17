package db2

import (
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKey(t *testing.T) {
	correctKey := Key("72b062c1-7fcf-11e7-ab09-acde48001122$42")

	assert.Equal(t, util.UID("72b062c1-7fcf-11e7-ab09-acde48001122"), correctKey.GetUID(), "Correct UID expected")
	assert.Equal(t, Generation(42), correctKey.GetGeneration(), "Correct Generation expected")

	noGenerationKey := Key("72b062c1-7fcf-11e7-ab09-acde48001122")

	assert.Panics(t, func() { noGenerationKey.GetUID() }, "Panic expected if key is incorrect")
	assert.Panics(t, func() { noGenerationKey.GetGeneration() }, "Panic expected if key is incorrect")

	invalidGenerationKey := Key("72b062c1-7fcf-11e7-ab09-acde48001122$bad")

	assert.Equal(t, util.UID("72b062c1-7fcf-11e7-ab09-acde48001122"), correctKey.GetUID(), "Correct UID expected")
	assert.Panics(t, func() { invalidGenerationKey.GetGeneration() }, "Panic expected if key is incorrect")
}
