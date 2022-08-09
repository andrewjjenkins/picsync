package googlephotos

type MediaItem struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	ProductUrl  string `json:"productUrl"`
	BaseUrl     string `json:"baseUrl"`
	MimeType    string `json:"mimeType"`
	Filename    string `json:"filename"`

	// MediaMetadata MediaMetadata `json:"mediaMetadata"`
	// ContributorInfo ContributorInfo `json:"contributorInfo"`
}
