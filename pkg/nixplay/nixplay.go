package nixplay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type loginResponse struct {
	Valid   bool `json:"valid"`
	Success bool `json:"success"`
	Errors  struct {
		All struct {
			Messages [][]string `json:"messages"`
		} `json:"__all__"`
	} `json:"errors"`
	Token string `json:"token"`
}

// Login logs in to nixplay
func Login(username string, password string) {
	resp, err := http.PostForm(
		"https://api.nixplay.com/www-login/",
		url.Values{
			"email":          {username},
			"password":       {password},
			"login_remember": {"true"},
			"undefined":      {"Log in"},
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	defer resp.Body.Close()
	fmt.Printf("Response: %v\n", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Body: %v\n", body)
	bodyStr := string(body[:])
	fmt.Printf("Bodystr: %v\n", bodyStr)

	authResult := &loginResponse{}
	err = json.Unmarshal(body, authResult)
	if err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
	}

	fmt.Printf("Unmarshalled response: %v\n", authResult)
}
