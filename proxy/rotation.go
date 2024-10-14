package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

var (
	errInvalidProxyUrl      = errors.New("invalid proxy URL")
	errUnsupportedProxyType = errors.New("unsupported proxy type")
	errCreatingSocks5Dialer = errors.New("error init SOCKS5 dialer")
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

func (r *Rotation) updateDialer() error {
	currentProxy := r.currentProxy()

	switch currentProxy.Type {
	case "http", "https":
		proxyURL, err := url.Parse(currentProxy.String())
		if err != nil {
			return errInvalidProxyUrl
		}
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return errInvalidProxyUrl
		}
		r.dialer = dialer
	case "socks5":
		dialer, err := proxy.SOCKS5("tcp", currentProxy.String(), nil, proxy.Direct)
		if err != nil {
			return errCreatingSocks5Dialer
		}
		r.dialer = dialer
	default:
		return errUnsupportedProxyType
	}
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

func (r *Rotation) Start(dur time.Duration) {
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for range ticker.C {
		if err := r.updateDialer(); err != nil {
			return
		}
		r.advanceProxyIndex()
		r.makeHttpClient()
	}
}

func (r *Rotation) currentProxy() Proxy {
	return r.proxies[r.currentProxyIdx]
}

func (r *Rotation) advanceProxyIndex() {
	r.currentProxyIdx = (r.currentProxyIdx + 1) % len(r.proxies)
}
