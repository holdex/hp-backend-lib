package liberr

type Error interface {
	error
	Code() string
}
