package nixplay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/andrewjjenkins/picsync/pkg/util"
)

type loginError struct {
	Messages [][]string `json:"messages"`
}

// Yes, unfortunately the shape of the response is different if there is a
// successful login or failure.
type loginResponseFailure struct {
	Valid   bool                  `json:"valid"`
	Success bool                  `json:"success"`
	Errors  map[string]loginError `json:"errors"`
	Token   string                `json:"token"`
}

type loginResponseSuccess struct {
	Valid   bool     `json:"valid"`
	Success bool     `json:"success"`
	Errors  []string `json:"errors"` // This is always an empty array
	Token   string   `json:"token"`
}

type auth struct {
	Token string
	Jar   http.CookieJar
}

// doLogin logs in to nixplay
func doLogin(username string, password string) (auth, error) {
	uStr := "https://api.nixplay.com/www-login/"
	u, err := url.Parse(uStr)
	if err != nil {
		return auth{}, err
	}
	resp, err := http.PostForm(
		uStr,
		url.Values{
			"email":          {username},
			"password":       {password},
			"login_remember": {"true"},
			"undefined":      {"Log in"},
		},
	)
	if err != nil {
		return auth{}, err
	}
	defer resp.Body.Close()

	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return auth{}, err
	}
	cookies := util.ReadSetCookies(resp.Header)
	for _, c := range cookies {
		if !strings.HasSuffix(c.Domain, ".nixplay.com") && c.Name != "AWSELB" {
			fmt.Printf("Skipping cookie %s, domain %s dangerous\n", c.Name, c.Domain)
			continue
		}
		jar.SetCookies(u, []*http.Cookie{c})
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return auth{}, err
	}

	authOk := &loginResponseSuccess{}
	err = json.Unmarshal(body, authOk)
	if err != nil {
		// Maybe it is a failure
		authFail := &loginResponseFailure{}
		err = json.Unmarshal(body, authFail)
		if err != nil {
			// Couldn't unmarshal as success or failure.
			return auth{}, err
		}

		// Parsed as error. There can be a map of an array of arrays of strings as
		// error messages but it looks like __all__.messages[0] is a "primary"
		// message.
		all, ok := authFail.Errors["__all__"]
		if !ok {
			return auth{}, errors.New("unknown error logging in")
		}
		if len(all.Messages) < 1 {
			return auth{}, errors.New("unknown error logging in")
		}
		msgs := all.Messages[0]
		if len(msgs) < 1 {
			return auth{}, errors.New("unknown error logging in")
		}
		return auth{}, errors.New(msgs[0])
	}
	return auth{
		Token: authOk.Token,
		Jar:   jar,
	}, nil
}

// Login logs in to nixplay
func Login(username string, password string) (*http.Client, error) {
	auth, err := doLogin(username, password)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
		Jar:     auth.Jar,
	}
	return client, nil
}

// GetConfig will get the user/app config
func GetConfig(c *http.Client) {
	resp, err := c.Get("https://api.nixplay.com/v2/app/config/")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	defer resp.Body.Close()
	fmt.Printf("Response: %v\n", resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Body: %v\n", string(body[:]))
}

func doPost(c *http.Client, urlString string, values *url.Values) (*http.Response, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		"POST",
		urlString,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")

	var csrfToken string
	cookies := c.Jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == "prod.csrftoken" {
			if csrfToken != "" {
				return nil, fmt.Errorf(
					"multiple Nixplay CSRF protection cookies (%s)",
					csrfToken,
				)
			}
			csrfToken = cookie.Value
		}
	}
	if csrfToken == "" {
		return nil, fmt.Errorf("no Nixplay CSRF protection cookie found")
	}
	req.Header.Set("X-CSRFToken", csrfToken)

	req.Header.Set("Origin", "https://app.nixplay.com")
	req.Header.Set("Referer", "https://app.nixplay.com/")

	return c.Do(req)
}
