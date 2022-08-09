package googlephotos

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GooglephotosAuth struct {
	Access oauth2.Token `json:"access,omitempty"`
}

func newOauth2Config(consumerKey string, consumerSecret string, redirectUrl string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     consumerKey,
		ClientSecret: consumerSecret,
		RedirectURL:  redirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/photoslibrary.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

func waitForCode(cc *CodeCatcher) string {
	for {
		select {
		case code := <-cc.Codes:
			// On success, give the server a chance to send some HTML
			// back to the user.  The login flow works even if this doesn't
			// happen, but it prevents the user from staring at an error
			// page in their browser.
			time.Sleep(2 * time.Second)
			cc.Server.Close()
			return code
		case err := <-cc.Errors:
			fmt.Printf("Error while waiting for code: %s\n", err.Error())
		}
	}
}

// Login does the OAuth2 login flow to google photos, resulting in Access tokens
func Login(consumerKey string, consumerSecret string) (*oauth2.Token, error) {

	// Google only allows OAuth2 via callback (even to localhost), it no longer
	// allows "OOB" OAuth2 flows (to mitigate phishing).  So we must start up a
	// web server to catch the code from the user's browser.
	codeCatcher, err := newCodeCatcher()
	if err != nil {
		return nil, err
	}

	config := newOauth2Config(consumerKey, consumerSecret, codeCatcher.CatcherURL)

	fmt.Printf("Follow this link to authorize:\n%s\n\n", config.AuthCodeURL(codeCatcher.State))
	code := waitForCode(codeCatcher)
	fmt.Printf("Successfully got one-time code from OAuth2 login, exchanging for tokens\n")
	token, err := config.Exchange(context.TODO(), code, oauth2.AccessTypeOffline)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func Client(consumerKey string, consumerSecret string, ctx context.Context, t *oauth2.Token) *http.Client {
	config := newOauth2Config(consumerKey, consumerSecret, "")
	return config.Client(ctx, t)
}
