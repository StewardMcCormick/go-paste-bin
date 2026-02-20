package error

type ValidationError struct {
	Message string                 `json:"message"`
	Status  int                    `json:"status"`
	Errors  []ValidationFieldError `json:"errors"`
}

type ValidationFieldError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

func (ve ValidationError) Error() string {
	return ve.Message
}

func (ve ValidationError) Code() int {
	return ve.Status
}
