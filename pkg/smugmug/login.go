package smugmug

import (
	"fmt"
	"net/http"

	"github.com/dghubble/oauth1"
)

const (
	requestTokenURL = "https://secure.smugmug.com/services/oauth/1.0a/getRequestToken"
	accessTokenURL  = "https://secure.smugmug.com/services/oauth/1.0a/getAccessToken"
	authorizeURL    = "https://secure.smugmug.com/services/oauth/1.0a/authorize"
)

// AccessAuth is the OAuth info required to get a new session token
type AccessAuth struct {
	Token          string
	Secret         string
	ConsumerKey    string
	ConsumerSecret string
}

func newOauthConfig(consumerKey string, consumerSecret string) *oauth1.Config {
	return &oauth1.Config{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: requestTokenURL,
			AuthorizeURL:    authorizeURL,
			AccessTokenURL:  accessTokenURL,
		},
		CallbackURL: "oob",
	}
}

// Login does the OAuth1 login flow to smugmug, resulting in Access tokens
func Login(consumerKey string, consumerSecret string) (*AccessAuth, error) {
	config := newOauthConfig(consumerKey, consumerSecret)

	reqToken, reqSecret, err := config.RequestToken()
	if err != nil {
		return nil, err
	}

	authURL, err := config.AuthorizationURL(reqToken)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Go to %s to authorize\n", authURL.String())
	fmt.Printf("Enter the six-digit code: ")
	var verifier string
	entries, err := fmt.Scanln(&verifier)
	if err != nil {
		return nil, err
	}
	if entries != 1 {
		return nil, fmt.Errorf("Did not read a verifier (%d)", entries)
	}

	accessToken, accessSecret, err := config.AccessToken(reqToken, reqSecret, verifier)

	return &AccessAuth{
		Token:  accessToken,
		Secret: accessSecret,
	}, nil
}

// Access returns a client that will use the access auth info to sign requests
func Access(access *AccessAuth) (*http.Client, error) {
	config := newOauthConfig(access.ConsumerKey, access.ConsumerSecret)
	token := oauth1.NewToken(access.Token, access.Secret)
	client := config.Client(oauth1.NoContext, token)
	req, err := http.NewRequest("GET", "https://api.smugmug.com/api/v2!authuser", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("login failed (%d)", resp.StatusCode)
	}
	return client, nil
}

/*
// Login logs in to SmugMug
func Login(consumerKey string, consumerSecret string) (*http.Client, error) {
	_, err := doLogin(consumerKey, consumerSecret)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
		//Jar:     auth.Jar,
	}
	return client, nil
}
*/
