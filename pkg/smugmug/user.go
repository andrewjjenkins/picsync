package smugmug

import (
	"fmt"
	"net/http"
)

// User is the metadata for a SmugMug user
type User struct {
	URI            string `json:"Uri"`
	URIDescription string `json:"UriDescription"`
	ResponseLevel  string
	AccountStatus  string
	Domain         string
	DomainOnly     string
	FirstName      string
	LastName       string
	Name           string
	NickName       string
	FriendsView    bool
	ImageCount     int
	IsTrial        bool
	Plan           string
	QuickShare     bool
	RefTag         string
	SortBy         string
	ViewPassHint   string
	ViewPassword   string
	WebURI         string            `json:"WebUri"`
	URIs           map[string]URIRef `json:"Uris"`
}

type userResponse struct {
	Response struct {
		responseCommon
		User User
	}
	Code    int
	Message string
}

func (u User) String() string {
	return fmt.Sprintf(
		"User %s (%s, domain %s): %s",
		u.NickName, u.Name, u.Domain, u.WebURI,
	)
}

// GetUser gets the metadata for a user
func GetUser(c *http.Client, nickname string) (*User, error) {
	resp := userResponse{}
	url := fmt.Sprintf("https://api.smugmug.com/api/v2/user/%s", nickname)
	err := GetUnmarshalJSON(c, url, &resp)
	return &resp.Response.User, err
}

// GetThisUser gets the user that you are authenticated as
func GetThisUser(c *http.Client) (*User, error) {
	resp := userResponse{}
	url := fmt.Sprintf("https://api.smugmug.com/api/v2!authuser")
	err := GetUnmarshalJSON(c, url, &resp)
	return &resp.Response.User, err
}
