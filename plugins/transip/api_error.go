package main

type ApiError struct {
	Code    int
	Message string
}

func (a ApiError) Error() string {
	return a.Message
}
