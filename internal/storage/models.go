package storage

type FileDataRequest struct {
	Key string `json:"key"`
}

type FileDataResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type UploadResponse struct {
	UploadUrl string `json:"uploadUrl"`
}
