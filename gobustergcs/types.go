package gobustergcs

// GCSError represents a returned error from GCS
type GCSError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Message      string `json:"message"`
			Reason       string `json:"reason"`
			LocationType string `json:"locationType"`
			Location     string `json:"location"`
		} `json:"errors"`
	} `json:"error"`
}

// GCSListing contains only a subset of returned properties
type GCSListing struct {
	IsTruncated string `json:"nextPageToken"`
	Items       []struct {
		Name         string `json:"name"`
		LastModified string `json:"updated"`
		Size         string `json:"size"`
	} `json:"items"`
}
