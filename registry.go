package serviceregistry

type Registry interface {
	Register(service string, protocol string, server string) error
	Get(service string, protocol string) ([]string, error)
	Close() error
}
