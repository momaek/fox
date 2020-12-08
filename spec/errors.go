package engine

// HTTPError engine error
type HTTPError interface {
	StatusCode() int
}
