package tracker

type KeepAliveRequest struct {
	ClientKey string `json:"client_key"`
	Address   string `json:"address"`
}

type ApiError struct {
	Error string `json:"error"`
}
