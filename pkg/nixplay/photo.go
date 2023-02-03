package nixplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/andrewjjenkins/picsync/pkg/util"
)

// Photo is a reference to a photo in an album
type Photo struct {
	OriginalURL     string `json:"original_url"`
	Orientation     int    `json:"orientation"`
	Caption         string `json:"caption"`
	URL             string `json:"url"`
	RotationAllowed bool   `json:"rotation_allowed"`
	Filename        string `json:"filename"` // Looks like the user's filename when uploaded
	SortDate        string `json:"sortDate"` // 20180731232531
	AlbumID         int    `json:"albumId"`
	S3Filename      string `json:"s3filename"`
	PreviewURL      string `json:"previewUrl"`
	Published       bool   `json:"published"`
	SourceID        string `json:"source_id"` // Literally "unused" for my examples
	Rotation        int    `json:"rotation"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	ID              int    `json:"id"`
	Md5             string `json:"md5"`
}

// GetPhotos returns the photos in an album
func (c *clientImpl) GetPhotos(albumID int, page int, limit int) ([]*Photo, error) {
	type getPhotosResponse struct {
		Photos []*Photo `json:"photos"`
	}
	photos := getPhotosResponse{}
	u := fmt.Sprintf(
		"https://api.nixplay.com/album/%d/pictures/json/?page=%d&limit=%d",
		albumID,
		page,
		limit,
	)
	err := util.GetUnmarshalJSON(c.httpClient, u, &photos)
	if err != nil {
		c.prom.getPhotosFailure.Inc()
	} else {
		c.prom.getPhotosSuccess.Inc()
		c.prom.getPhotosPhotoCount.Add(float64(len(photos.Photos)))
	}
	return photos.Photos, err
}

func (p Photo) String() string {
	return fmt.Sprintf(
		"%s (%d/%d): %s [%s]",
		p.Filename, p.AlbumID, p.ID, p.S3Filename, p.Caption,
	)
}

type recResponse struct {
	Token string `json:"token"`
}

func getUploadToken(c *http.Client, albumID int) (string, error) {
	vals := url.Values{
		"albumId": {fmt.Sprintf("%d", albumID)},
		"total":   {"1"},
	}
	resp, err := doPost(c, "https://api.nixplay.com/v3/upload/receivers/", &vals)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("couldn't create receiver: %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	recResp := recResponse{}
	err = json.Unmarshal(body, &recResp)
	if err != nil {
		return "", err
	}
	return recResp.Token, nil
}

type uploadVals struct {
	Token    string
	AlbumID  int
	FileName string
	FileType string
	FileSize uint64
}

type uploader struct {
	Data struct {
		ACL            string `json:"acl"`
		Key            string `json:"key"`
		AWSAccessKeyID string `json:"AWSAccessKeyId"`
		Policy         string
		Signature      string
		UserUploadID   string   `json:"userUploadId"`
		BatchUploadID  string   `json:"batchUploadId"`
		UserUploadIds  []string `json:"userUploadIds"`
		FileType       string   `json:"fileType"`
		FileSize       int      `json:"fileSize"`
		S3UploadURL    string   `json:"s3UploadUrl"`
	} `json:"data"`
}

func getUploader(c *http.Client, v uploadVals) (*uploader, error) {
	vals := url.Values{
		"uploadToken": {v.Token},
		"albumId":     {fmt.Sprintf("%d", v.AlbumID)},
		"fileName":    {v.FileName},
		"fileType":    {v.FileType},
		"fileSize":    {fmt.Sprintf("%d", v.FileSize)},
	}
	resp, err := doPost(c, "https://api.nixplay.com/v3/photo/upload/", &vals)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("couldn't create uploader: %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	upResp := uploader{}
	err = json.Unmarshal(body, &upResp)
	if err != nil {
		return nil, err
	}
	return &upResp, nil
}

func uploadS3(u *uploader, filename string, body io.ReadCloser) error {

	reqBody := &bytes.Buffer{}
	writer := multipart.NewWriter(reqBody)

	formVals := map[string]string{
		"key":                        u.Data.Key,
		"acl":                        u.Data.ACL,
		"content-type":               u.Data.FileType,
		"x-amz-meta-batch-upload-id": u.Data.BatchUploadID,
		"success_action_status":      "201",
		"AWSAccessKeyId":             u.Data.AWSAccessKeyID,
		"Policy":                     u.Data.Policy,
		"Signature":                  u.Data.Signature,
	}
	for k, v := range formVals {
		toWrite, err := writer.CreateFormField(k)
		if err != nil {
			return err
		}
		io.WriteString(toWrite, v)
	}

	fileWrite, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(fileWrite, body)
	if err != nil {
		return err
	}
	body.Close()
	writer.Close()

	req, err := http.NewRequest(
		"POST",
		u.Data.S3UploadURL,
		reqBody,
	)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	ct := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
	req.Header.Set("content-type", ct)
	req.Header.Set("origin", "https://app.nixplay.com")
	req.Header.Set("referer", "https://app.nixplay.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return fmt.Errorf("error uploading: %s", resp.Status)
	}
	return nil
}

// UploadPhoto uploads a photo to an album
func (c *clientImpl) UploadPhoto(albumID int, filename string, filetype string, filesize uint64, body io.ReadCloser) error {
	uploadToken, err := getUploadToken(c.httpClient, albumID)
	if err != nil {
		c.prom.uploadPhotoFailure.Inc()
		return err
	}

	uploader, err := getUploader(
		c.httpClient,
		uploadVals{
			Token:    uploadToken,
			AlbumID:  albumID,
			FileName: filename,
			FileType: filetype,
			FileSize: filesize,
		},
	)
	if err != nil {
		c.prom.uploadPhotoFailure.Inc()
		return err
	}

	err = uploadS3(uploader, filename, body)
	if err != nil {
		c.prom.uploadPhotoFailure.Inc()
	} else {
		c.prom.uploadPhotoSuccess.Inc()
		c.prom.uploadPhotoTotalBytes.Add(float64(filesize))
	}
	return err
}

func (c *clientImpl) DeletePhoto(id int) error {
	u := fmt.Sprintf("https://api.nixplay.com/picture/%d/delete/json/", id)
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		c.prom.deletePhotoFailure.Inc()
		return err
	}
	req.Header.Set("accept", "application/json")
	res, err := doNixplayCsrf(c.httpClient, req)
	if err != nil {
		c.prom.deletePhotoFailure.Inc()
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		c.prom.deletePhotoFailure.Inc()
		return fmt.Errorf("received http %d when deleting nixplay photo %d",
			res.StatusCode, id)
	}
	c.prom.deletePhotoSuccess.Inc()
	// We don't care about what's in the body; 200 OK is good enough
	return nil
}
