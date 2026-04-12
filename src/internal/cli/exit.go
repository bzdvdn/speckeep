package cli

type exitError struct {
	code    int
	message string
}

func (e *exitError) Error() string {
	return e.message
}

func (e *exitError) ExitCode() int {
	return e.code
}

func newExitError(code int, message string) error {
	return &exitError{code: code, message: message}
}
