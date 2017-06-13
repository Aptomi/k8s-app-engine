package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadSecrets(t *testing.T) {
	secrets := LoadUserSecretsByIDFromDir("testdata/unittests", "1")

	assert.Equal(t, 4, len(secrets))
	assert.Equal(t, "aliceappkey", secrets["twitterAppKey"])
	assert.Equal(t, "aliceappsecret", secrets["twitterAppSecret"])
	assert.Equal(t, "alicetokenkey", secrets["twitterTokenKey"])
	assert.Equal(t, "alicetokensecret", secrets["twitterTokenSecret"])

	secrets = LoadUserSecretsByIDFromDir("testdata/unittests", "2")

	assert.Equal(t, 4, len(secrets))
	assert.Equal(t, "bobappkey", secrets["twitterAppKey"])
	assert.Equal(t, "bobappsecret", secrets["twitterAppSecret"])
	assert.Equal(t, "bobtokenkey", secrets["twitterTokenKey"])
	assert.Equal(t, "bobtokensecret", secrets["twitterTokenSecret"])

	secrets = LoadUserSecretsByIDFromDir("testdata/unittests", "3")
	assert.Equal(t, 1, len(secrets))
	assert.Equal(t, "topsecret", secrets["someSecret"])

	secrets = LoadUserSecretsByIDFromDir("testdata/unittests", "4")
	assert.Equal(t, 0, len(secrets))
}

func TestUserLabelsWithSecrets(t *testing.T) {
	userAlice := LoadUserByIDFromDir("testdata/unittests", "1")
	labels := userAlice.getLabelSet()

	assert.Equal(t, 9, len(labels.Labels))
	assert.Equal(t, "aliceappkey", labels.Labels["twitterAppKey"])
	assert.Equal(t, "platform_services", labels.Labels["team"])
}

func TestUserLabelsWithEmptySecrets(t *testing.T) {
	userDave := LoadUserByIDFromDir("testdata/unittests", "5")
	labels := userDave.getLabelSet()

	assert.Equal(t, 5, len(labels.Labels))
}
