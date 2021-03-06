package clarifai

import (
	"encoding/json"
	"errors"
	"math/big"
)

// InfoResp represents the expected JSON response from /info/
type InfoResp struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_msg"`
	Results       struct {
		MaxImageSize      int     `json:"max_image_size"`
		DefaultLanguage   string  `json:"default_language"`
		MaxVideoSize      int     `json:"max_video_size"`
		MaxImageBytes     int     `json:"max_image_bytes"`
		DefaultModel      string  `json:"default_model"`
		MaxVideoBytes     int     `json:"max_video_bytes"`
		MaxVideoDuration  int     `json:"max_video_duration"`
		MaxVideoBatchSize int     `json:"max_video_batch_size"`
		MinVideoSize      int     `json:"min_video_size"`
		MinImageSize      int     `json:"min_image_size"`
		MaxBatchSize      int     `json:"max_batch_size"`
		APIVersion        float32 `json:"api_version"`
	}
}

// ColorRequest represents a JSON request for /color/
type ColorRequest struct {
	URLs  []string `json:"url"`
	Files []string `json:"files,omitempty"`
}

// GetFiles returns the files on the request.
func (c ColorRequest) GetFiles() []string {
	return c.Files
}

// GetModel returns the model for the request.
func (c ColorRequest) GetModel() string {
	return ""
}

// ColorResp represents the expected JSON response from /color/
type ColorResp struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_msg"`
	Results       []*ColorResult
}

// ColorResult represents the expected data for a single color result
type ColorResult struct {
	DocID       *big.Int `json:"docid"`
	URL         string   `json:"url"`
	Colors      []*Color `json:"colors"`
	DocIDString string   `json:"docid_str"`
}

// Color represents a color.
type Color struct {
	Hex     string  `json:"hex"`
	Density float64 `json:"density"`
	W3C     W3C     `json:"w3c"`
}

// W3C is to reprsent the W3C color.
type W3C struct {
	Hex  string `json:"hex"`
	Name string `json:"name"`
}

// TagRequest represents a JSON request for /tag/
type TagRequest struct {
	URLs     []string `json:"url"`
	Files    []string `json:"files,omitempty"`
	LocalIDs []string `json:"local_ids,omitempty"`
	Model    string   `json:"model,omitempty"`
}

// GetFiles returns the files on the request.
func (t TagRequest) GetFiles() []string {
	return t.Files
}

// GetModel returns the model for the request.
func (t TagRequest) GetModel() string {
	return t.Model
}

// TagResp represents the expected JSON response from /tag/
type TagResp struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_msg"`
	Meta          struct {
		Tag struct {
			Timestamp json.Number `json:"timestamp"`
			Model     string      `json:"model"`
			Config    string      `json:"config"`
		}
	}
	Results []TagResult
}

// TagResult represents the expected data for a single tag result
type TagResult struct {
	DocID         *big.Int `json:"docid"`
	URL           string   `json:"url"`
	StatusCode    string   `json:"status_code"`
	StatusMessage string   `json:"status_msg"`
	LocalID       string   `json:"local_id"`
	Result        struct {
		Tag struct {
			Classes []string  `json:"classes"`
			CatIDs  []string  `json:"catids"`
			Probs   []float32 `json:"probs"`
		}
	}
	DocIDString string `json:"docid_str"`
}

// FeedbackForm is used to send feedback back to Clarifai
type FeedbackForm struct {
	DocIDs           []string `json:"docids,omitempty"`
	URLs             []string `json:"url,omitempty"`
	AddTags          []string `json:"add_tags,omitempty"`
	RemoveTags       []string `json:"remove_tags,omitempty"`
	DissimilarDocIDs []string `json:"dissimilar_docids,omitempty"`
	SimilarDocIDs    []string `json:"similar_docids,omitempty"`
	SearchClick      []string `json:"search_click,omitempty"`
}

// FeedbackResp is the expected response from /feedback/
type FeedbackResp struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_msg"`
}

type hasFiles interface {
	GetFiles() []string
	GetModel() string
}

// Info will return the current status info for the given client
func (client *Client) Info() (*InfoResp, error) {
	res, err := client.commonHTTPRequest(nil, "info", "GET", false)

	if err != nil {
		return nil, err
	}

	info := new(InfoResp)
	err = json.Unmarshal(res, info)

	return info, err
}

// Tag allows the client to request tag data on a single, or multiple photos
func (client *Client) Tag(req TagRequest) (*TagResp, error) {
	if len(req.URLs) < 1 && len(req.Files) < 1 {
		return nil, errors.New("Requires at least one file or url")
	}

	// API doesn't support file and URLs simultaniously.
	if len(req.Files) > 0 && len(req.URLs) > 0 {
		return nil, errors.New("Can't submit both files and urls")
	}

	res := []byte{}
	var err error

	if len(req.Files) > 0 {
		res, err = client.fileHTTPRequest(req, "tag", false)
	} else {
		res, err = client.commonHTTPRequest(req, "tag", "POST", false)
	}

	if err != nil {
		return nil, err
	}

	tagres := new(TagResp)
	err = json.Unmarshal(res, tagres)

	return tagres, err
}

// Color allows the client to request color data on a single, or multiple photos
func (client *Client) Color(req ColorRequest) (*ColorResp, error) {
	if len(req.URLs) < 1 && len(req.Files) < 1 {
		return nil, errors.New("Requires at least one file or url")
	}

	// API doesn't support file and URLs simultaniously.
	if len(req.Files) > 0 && len(req.URLs) > 0 {
		return nil, errors.New("Can't submit both files and urls")
	}

	res := []byte{}
	var err error

	if len(req.Files) > 0 {
		res, err = client.fileHTTPRequest(req, "color", false)
	} else {
		res, err = client.commonHTTPRequest(req, "color", "POST", false)
	}

	if err != nil {
		return nil, err
	}

	colorRes := new(ColorResp)
	err = json.Unmarshal(res, colorRes)

	return colorRes, err
}

// Feedback allows the user to provide contextual feedback to Clarifai in order to improve their results
func (client *Client) Feedback(form FeedbackForm) (*FeedbackResp, error) {
	if form.DocIDs == nil && form.URLs == nil {
		return nil, errors.New("Requires at least one docid or url")
	}

	if form.DocIDs != nil && form.URLs != nil {
		return nil, errors.New("Request must provide exactly one of the following fields: {'DocIDs', 'URLs'}")
	}

	res, err := client.commonHTTPRequest(form, "feedback", "POST", false)

	feedbackres := new(FeedbackResp)
	err = json.Unmarshal(res, feedbackres)

	return feedbackres, err

}
