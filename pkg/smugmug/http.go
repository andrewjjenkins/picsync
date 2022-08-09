package smugmug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// URIRef is the SmugMug API type for a URI in a response
type URIRef struct {
	URI            string `json:"Uri"`
	Locator        string
	LocatorType    string
	URIDescription string `json:"UriDescription"`
	EndpointType   string
}

type responseCommon struct {
	URI            string `json:"Uri"`
	URIDescription string `json:"UriDescription"`
	DocURI         string `json:"DocUri"`
	EndpointType   string
	Locator        string
	LocatorType    string
}

// GetUnmarshalJSON gets a JSON response from url and unmarshals into target
func GetUnmarshalJSON(c *http.Client, url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}
