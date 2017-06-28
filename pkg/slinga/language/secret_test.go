package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadSecrets(t *testing.T) {
	secrets := LoadUserSecretsByIDFromDir("../testdata/unittests", "1")

	assert.Equal(t, 4, len(secrets))
	assert.Equal(t, "aliceappkey", secrets["twitterAppKey"])
	assert.Equal(t, "aliceappsecret", secrets["twitterAppSecret"])
	assert.Equal(t, "alicetokenkey", secrets["twitterTokenKey"])
	assert.Equal(t, "alicetokensecret", secrets["twitterTokenSecret"])

	secrets = LoadUserSecretsByIDFromDir("../testdata/unittests", "2")

	assert.Equal(t, 4, len(secrets))
	assert.Equal(t, "bobappkey", secrets["twitterAppKey"])
	assert.Equal(t, "bobappsecret", secrets["twitterAppSecret"])
	assert.Equal(t, "bobtokenkey", secrets["twitterTokenKey"])
	assert.Equal(t, "bobtokensecret", secrets["twitterTokenSecret"])

	secrets = LoadUserSecretsByIDFromDir("../testdata/unittests", "3")
	assert.Equal(t, 1, len(secrets))
	assert.Equal(t, "topsecret", secrets["someSecret"])

	secrets = LoadUserSecretsByIDFromDir("../testdata/unittests", "4")
	assert.Equal(t, 0, len(secrets))
}

func TestUserWithSecrets(t *testing.T) {
	userAlice := NewUserLoaderFromDir("../testdata/unittests").LoadUserByID("1")
	secrets := userAlice.GetSecretSet()

	assert.Equal(t, 4, len(secrets.Labels))
	assert.Equal(t, "aliceappkey", secrets.Labels["twitterAppKey"])
}

func TestUserWithEmptySecrets(t *testing.T) {
	userDave := NewUserLoaderFromDir("../testdata/unittests").LoadUserByID("5")
	secrets := userDave.GetSecretSet()

	assert.Equal(t, 0, len(secrets.Labels))
}
