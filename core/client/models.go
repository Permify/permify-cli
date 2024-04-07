package client

// ErrorResponse defines the generic error response received from the server
type ErrorResponse struct {
	StatusCode	int  		`json:"status_code"`
	ErrorCode	int			`json:"code"`
	Message		string 		`json:"message"`
	Detail		[]string	`json:"detail"` 
}