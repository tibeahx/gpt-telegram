package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type Rotation struct {
	proxies         []Proxy
	currentProxyIdx int
	dialer          proxy.Dialer
	httpClient      *http.Client
}

func NewRotation(filepath string) (*Rotation, error) {
	var p Proxy
	proxies, err := p.FromFile(filepath)
	if err != nil {
		return &Rotation{}, err
	}
	rotate := &Rotation{
		proxies: proxies,
	}
	return rotate, nil
}

var errInvalidProxyUrl = errors.New("invalid proxy URL")

func (r *Rotation) updateDialer() error {
	cp := r.currentProxy()
	proxyURL, err := url.Parse(cp.String())
	if err != nil {
		return errInvalidProxyUrl
	}
	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return errInvalidProxyUrl
	}
	r.dialer = dialer
	return nil
}

func (r *Rotation) makeHttpClient() {
	httpTransport := &http.Transport{
		Dial: r.dialer.Dial,
	}
	r.httpClient = &http.Client{
		Transport: httpTransport,
	}
}

func (r *Rotation) Start(dur time.Duration, wg *sync.WaitGroup) {
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for range ticker.C {
		wg.Add(1)
		if err := r.updateDialer(); err != nil {
			return
		}
		r.advanceProxyIndex()
		r.makeHttpClient()
		wg.Done()
	}
}

func (r *Rotation) currentProxy() Proxy {
	return r.proxies[r.currentProxyIdx]
}

func (r *Rotation) advanceProxyIndex() {
	r.currentProxyIdx = (r.currentProxyIdx + 1) % len(r.proxies)
}
