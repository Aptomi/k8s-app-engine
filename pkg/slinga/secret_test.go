package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadSecrets(t *testing.T) {
	secrets := LoadUserSecretsByIDFromDir("testdata/unittests", "1")

	assert.Equal(t, 4, len(secrets.Labels))
	assert.Equal(t, "aliceappkey", secrets.Labels["twitterAppKey"])
	assert.Equal(t, "aliceappsecret", secrets.Labels["twitterAppSecret"])
	assert.Equal(t, "alicetokenkey", secrets.Labels["twitterTokenKey"])
	assert.Equal(t, "alicetokensecret", secrets.Labels["twitterTokenSecret"])

	secrets = LoadUserSecretsByIDFromDir("testdata/unittests", "2")

	assert.Equal(t, 4, len(secrets.Labels))
	assert.Equal(t, "bobappkey", secrets.Labels["twitterAppKey"])
	assert.Equal(t, "bobappsecret", secrets.Labels["twitterAppSecret"])
	assert.Equal(t, "bobtokenkey", secrets.Labels["twitterTokenKey"])
	assert.Equal(t, "bobtokensecret", secrets.Labels["twitterTokenSecret"])

	secrets = LoadUserSecretsByIDFromDir("testdata/unittests", "3")
	assert.Equal(t, 1, len(secrets.Labels))
	assert.Equal(t, "topsecret", secrets.Labels["someSecret"])

	secrets = LoadUserSecretsByIDFromDir("testdata/unittests", "4")
	assert.Equal(t, 0, len(secrets.Labels))
}

func TestUserLabelsWithSecrets(t *testing.T) {
	userAlice := LoadUserByIDFromDir("testdata/unittests", "1")
	labels := userAlice.getLabelSet("testdata/unittests")

	assert.Equal(t, 9, len(labels.Labels))
	assert.Equal(t, "aliceappkey", labels.Labels["twitterAppKey"])
	assert.Equal(t, "platform_services", labels.Labels["team"])
}

func TestUserLabelsWithEmptySecrets(t *testing.T) {
	userDave := LoadUserByIDFromDir("testdata/unittests", "5")
	labels := userDave.getLabelSet("testdata/unittests")

	assert.Equal(t, 5, len(labels.Labels))
}
