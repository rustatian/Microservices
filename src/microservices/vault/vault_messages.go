package vault

type hashRequest struct {
	Password string `json:"password"`
}

type hashResponse struct {
	Hash string `json:"hash"`
	Err  string `json:"err, omitempty"`
}

type validateRequest struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}

type validateResponse struct {
	Valid bool   `json:"valid"`
	Err   string `json:"err, omitempty"`
}

//TODO Create health logic, check free memory, disk space
type healthRequest struct{}

type healthResponse struct {
	Status bool `json:"status"`
}
