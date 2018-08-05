package smugmug

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/andrewjjenkins/nixplay/pkg/util"
	"golang.org/x/net/publicsuffix"
)

type loginRequest struct {
	Email        string `json:"Email"`
	Password     string `json:"Password"`
	OTPCode      string `json:"OTPCode"` // Empty for me, but still present
	KeepLoggedIn int    `json:"KeepLoggedIn"`
	IsOAuth      int    `json:"IsOAuth"`
	Method       string `json:"method"`
	Token        string `json:"_token"`
}

type auth struct {
	Jar http.CookieJar
}

type loginSession struct {
	String    string `json:"string"` // https://www.smugmug.com
	Time      int    `json:"time"`   //
	Signature string `json:"signature"`
	Version   int    `json:"version"`
	Algorithm string `json:"algorithm"` // sha1
}

func getLoginSession() {

}

func doLogin(username string, password string) (*auth, error) {
	uStr := "https://secure.smugmug.com/services/api/json/1.4.0/"
	u, err := url.Parse(uStr)
	if err != nil {
		return nil, err
	}

	loginVals := url.Values{
		"Email":        {username},
		"Password":     {password},
		"OTPCode":      {""},
		"KeepLoggedIn": {"1"},
		"IsOAuth":      {"0"},
		"method":       {"rpc.user.login"},
		"_token":       {""},
	}

	req, err := http.NewRequest("POST", uStr, strings.NewReader(loginVals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("origin", "https://secure.smugmug.com")
	req.Header.Add("referer", "https://secure.smugmug.com/login")
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error login: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}
	cookies := util.ReadSetCookies(resp.Header)
	for _, c := range cookies {
		if !strings.HasSuffix(c.Domain, ".smugmug.com") {
			fmt.Printf("Skipping cookie %s, domain %s dangerous\n", c.Name, c.Domain)
			continue
		}
		jar.SetCookies(u, []*http.Cookie{c})
	}

	fmt.Printf("Response: %v\n", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Body: %v\n", body)
	bodyStr := string(body[:])
	fmt.Printf("Bodystr: %v\n", bodyStr)

	return nil, err
}

// Login logs in to SmugMug
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
