package network

import (
	"errors"
	"net"
)

type FilteredListener struct {
	allowedAddresses []string
	listener         net.Listener
}

var ErrUnauthorizedSource error = errors.New("unauthorized source")

func NewFilteredListener(allowed []string, listener net.Listener) (*FilteredListener, error) {
	return &FilteredListener{
		allowedAddresses: allowed,
		listener:         listener,
	}, nil
}

func (l *FilteredListener) Accept() (net.Conn, error) {
	c, err := l.listener.Accept()

	if err != nil {
		return nil, err
	}

	if !l.remoteAllowed(c.RemoteAddr()) {
		c.Close()
		return nil, ErrUnauthorizedSource
	}

	return c, nil
}

func (l *FilteredListener) Close() error {
	return l.listener.Close()
}

func (l *FilteredListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *FilteredListener) remoteAllowed(remote net.Addr) bool {
	remoteStr := remote.String()

	remoteIP, _, err := net.SplitHostPort(remoteStr)
	if err != nil {
		return false
	}

	for _, allowed := range l.allowedAddresses {
		if remoteIP == allowed {
			return true
		}
	}

	return false
}
