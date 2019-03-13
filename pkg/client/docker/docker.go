package docker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/distribution/registry/client/auth"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/transport"
)

const (
	defaultTimeOut = time.Minute * 5
	authClientID   = "ivm-controller"
)

var _ Interface = &Client{}

// Interface provides functionality to deal with container image tags.
type Interface interface {
	Tags(auth *Auth, repository string) ([]string, error)
}

// Auth is a helper to store authentication details for the client.
type Auth struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

func (c *Auth) withDefaults() {
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}
}

// Client is the docker implemention of Interface.
type Client struct{}

// Tags retrieves docker tags for a specific repository.
func (c *Client) Tags(authentication *Auth, repository string) ([]string, error) {

	if authentication == nil {
		authentication = &Auth{}
	}
	authentication.withDefaults()

	namedRef, err := reference.ParseNormalizedNamed(repository)

	if err != nil {
		return nil, err
	}

	imageName, err := reference.WithName(reference.Path(namedRef))

	if err != nil {
		return nil, err
	}

	modifiers := []transport.RequestModifier{transport.NewHeaderRequestModifier(http.Header{"User-Agent": []string{authClientID}})}
	authTransport := transport.NewTransport(authentication.Transport, modifiers...)

	registryURL, err := getRegistryURL(namedRef)

	if err != nil {
		return nil, err
	}

	challengeManager, _, err := PingV2Registry(registryURL, authTransport)

	if err != nil {
		return nil, err
	}

	scope := auth.RepositoryScope{
		Repository: imageName.Name(),
		Actions:    []string{"pull"},
	}

	creds := NewCredentialStore(authentication.Username, authentication.Password)

	tokenHandlerOptions := auth.TokenHandlerOptions{
		Transport:   authTransport,
		Credentials: creds,
		Scopes:      []auth.Scope{scope},
		ClientID:    "docker",
	}
	tokenHandler := auth.NewTokenHandlerWithOptions(tokenHandlerOptions)
	basicHandler := auth.NewBasicHandler(creds)

	modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler))

	tr := transport.NewTransport(authentication.Transport, modifiers...)

	repo, err := client.NewRepository(imageName, registryURL.String(), tr)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeOut)
	defer cancel()
	return repo.Tags(ctx).All(ctx)
}

func getRegistryURL(ref reference.Named) (*url.URL, error) {
	if domain := reference.Domain(ref); domain != "" {
		if domain == "docker.io" {
			domain = fmt.Sprintf("registry-1.%s", domain)
		}
		u, err := url.Parse(domain)
		if err != nil {
			return nil, err
		}
		u.Scheme = "https"
		return u, nil
	}
	return nil, errors.New("missing domain from image name")
}

// Split takes a string consisting of image name and tag and splits them into two.
func Split(image string) (base string, tag string, err error) {
	if image == "" {
		return "", "", errors.New("image can't be empty")
	}
	imageParts := strings.Split(image, ":")

	if len(imageParts) != 2 {
		return "", "", errors.New("invalid image format")
	}

	return imageParts[0], imageParts[1], nil
}
