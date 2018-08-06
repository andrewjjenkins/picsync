package smugmug

import (
	"fmt"
	"net/http"
)

// Node is the metadata for a SmugMug node (container).
type Node struct {
	URI                   string `json:"Uri"`
	URIDescription        string `json:"UriDescription"`
	ResponseLevel         string
	DateModified          string
	Description           string
	EffectiveSecurityType string
	HasChildren           bool
	IsRoot                bool
	Keywords              []string
	Name                  string
	NodeID                string
	PasswordHint          string
	SecurityType          string
	ShowCoverImage        bool
	SortDirection         string
	SortMethod            string
	Type                  string
	URLName               string            `json:"UrlName"`
	URLPath               string            `json:"UrlPath"`
	WebURI                string            `json:"WebUri"`
	URIs                  map[string]URIRef `json:"Uris"`

	// FormattedValues - skip
}

type nodeResponse struct {
	Response struct {
		responseCommon
		Node Node
	}
	Code    int
	Message string
}

func (n Node) String() string {
	return fmt.Sprintf("Node %s: (%s) %s", n.Name, n.NodeID, n.WebURI)
}

// GetNode will get the metadata for a node from the API
func GetNode(c *http.Client, id string) (*Node, error) {
	resp := nodeResponse{}
	url := fmt.Sprintf("https://api.smugmug.com/api/v2/node/%s", id)
	err := GetUnmarshalJSON(c, url, &resp)
	return &resp.Response.Node, err
}
