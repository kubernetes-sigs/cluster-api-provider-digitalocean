package machine

import (
	"context"
	"os"

	"github.com/digitalocean/godo"
	"github.com/golang/glog"

	"golang.org/x/oauth2"
)

// tokenSource contains API token for DigitalOcean API.
type tokenSource struct {
	AccessToken string
}

// Token returns new oauth2 object with DO API token.
func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// getGodoClient creates new godo client used to interact with the DigitalOcean API.
func getGodoClient() *godo.Client {
	doToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if doToken == "" {
		glog.Fatalf("env var DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	token := &tokenSource{
		AccessToken: doToken,
	}

	oc := oauth2.NewClient(context.Background(), token)
	return godo.NewClient(oc)
}
