package error

// CustomError — ошибка с HTTP-кодом для возврата клиенту.
type CustomError struct {
	Message    string
	StatusCode int
}

func (e CustomError) Error() string {
	return e.Message
}
