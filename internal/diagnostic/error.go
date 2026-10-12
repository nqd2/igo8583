/* Package diagnostic keeps payload-bearing causes out of default error strings. */
package diagnostic

type errorWithCause struct {
	message string
	cause   error
}

func (e *errorWithCause) Error() string { return e.message }
func (e *errorWithCause) Unwrap() error { return e.cause }

/* Redact retains the cause for errors.Is and errors.As without formatting it. */
func Redact(message string, cause error) error {
	return &errorWithCause{message: message, cause: cause}
}
