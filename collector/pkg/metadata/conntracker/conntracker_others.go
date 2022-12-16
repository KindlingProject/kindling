//go:build !linux

package conntracker

func NewConntracker(config *Config) (Conntracker, error) {
	return NewNoopConntracker(config), nil
}
