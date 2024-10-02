package tracker

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TrackerServer struct {
	listenAddress  string
	upstreams      map[string]*Upstream
	clientDeadline time.Duration

	serverCert string

	upstreamLock *sync.RWMutex
	upstreamRR   *UpstreamRoundRobin
}

type TrackerContext struct {
	echo.Context

	server *TrackerServer
}

var ErrNoUpstreamsAvailable = errors.New("no upstreams available")
var ErrNoSuchUpstream = errors.New("invalid upstream key")

func NewTrackerServer(listenAddress string, clientKeys []string, clientDeadline time.Duration, certFile string) *TrackerServer {
	upstreams := make(map[string]*Upstream)

	for _, key := range clientKeys {
		upstreams[key] = &Upstream{
			Address:   "",
			Key:       key,
			KeepAlive: time.Now(),
			Enabled:   true,
			Available: false,
		}
	}

	return &TrackerServer{
		listenAddress:  listenAddress,
		upstreams:      upstreams,
		clientDeadline: clientDeadline,
		upstreamLock:   &sync.RWMutex{},
		upstreamRR:     NewUpstreamRoundRobin(nil),

		serverCert: certFile,
	}
}

func (ts *TrackerServer) Run() error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	//e.Use(middleware.Logger())
	e.Use(ContextMiddleware(ts))

	e.POST("/api/keepalive", withContext(upstreamKeepAlive))
	e.GET("/api/upstreams", withContext(getUpstreams))

	return e.StartTLS(ts.listenAddress, ts.serverCert, ts.serverCert)

	//return e.Start(ts.listenAddress)
}

func (ts *TrackerServer) GetUpstream() (*Upstream, error) {
	for {
		ts.upstreamLock.RLock()
		// cannot defer unlock due to the following conditions

		if ts.upstreamRR.Len() == 0 {
			ts.upstreamLock.RUnlock()
			return nil, ErrNoUpstreamsAvailable
		}

		upstream := ts.upstreamRR.Next()
		ts.upstreamLock.RUnlock()

		if upstream.IsAlive(ts.clientDeadline) {
			return upstream, nil
		} else {
			// this call needs the mutex to be unlocked
			ts.removeAvailableUpstream(upstream)
		}
	}
}

func (ts *TrackerServer) UpdateUpstreamKeepalive(upstreamKey string, address string) error {
	upstream, ok := ts.upstreams[upstreamKey]
	if !ok {
		return ErrNoSuchUpstream
	}

	upstream.KeepAlive = time.Now()
	upstream.Address = address

	if !upstream.Available {
		ts.addAvailableUpstream(upstream)
	}

	return nil
}

func (ts *TrackerServer) addAvailableUpstream(upstream *Upstream) {
	ts.upstreamLock.Lock()
	defer ts.upstreamLock.Unlock()

	// another routine could have entered here between the previous check and now
	if upstream.Available {
		return
	}

	log.Printf("upstream %s is now available at %s!\n", upstream.Key, upstream.Address)

	ts.upstreamRR.Add(upstream)
	upstream.Available = true

}

func (ts *TrackerServer) removeAvailableUpstream(upstream *Upstream) {
	ts.upstreamLock.Lock()
	defer ts.upstreamLock.Unlock()

	// another routine could have entered here between previous check and now
	if !upstream.Available {
		return
	}

	log.Printf("upstream %s is no longer available!\n", upstream.Key)

	ts.upstreamRR.Remove(upstream)
	upstream.Available = false
}

func withContext(f func(TrackerContext) error) func(echoCtx echo.Context) error {
	return func(c echo.Context) error {
		tctx, ok := c.(TrackerContext)
		if !ok {
			return c.JSON(http.StatusInternalServerError, ApiError{"internal error"})
		}

		return f(tctx)
	}
}

func ContextMiddleware(ts *TrackerServer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := TrackerContext{
				Context: c,
				server:  ts,
			}
			return next(cc)
		}
	}
}

func upstreamKeepAlive(c TrackerContext) error {
	var request KeepAliveRequest

	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ApiError{""})
	}

	err = c.server.UpdateUpstreamKeepalive(request.ClientKey, request.Address)

	if err != nil {
		return c.JSON(http.StatusBadRequest, ApiError{err.Error()})
	}

	return c.JSON(http.StatusOK, ApiError{})
}

func getUpstreams(c TrackerContext) error {
	c.server.upstreamLock.RLock()
	defer c.server.upstreamLock.RUnlock()

	return c.JSON(http.StatusOK, c.server.upstreams)
}
