package docker

import (
	"net/url"

	"github.com/docker/distribution/registry/client/auth"
)

type credentialStore struct {
	username string
	password string
}

// NewCredentialStore provides static username and password to the store.
func NewCredentialStore(username, password string) auth.CredentialStore {
	return credentialStore{
		username: username,
		password: password,
	}
}

func (cs credentialStore) Basic(*url.URL) (string, string) {
	return cs.username, cs.password
}

func (cs credentialStore) RefreshToken(*url.URL, string) string {
	return ""
}

func (cs credentialStore) SetRefreshToken(*url.URL, string, string) {
}
