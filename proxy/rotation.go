package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
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
	proxies, err := p.FromFile(path.Join(".", filepath))
	if err != nil {
		return &Rotation{}, err
	}
	rotate := &Rotation{
		proxies: proxies,
	}
	return rotate, nil
}

func (r *Rotation) updateDialer() error {
	cp := r.currentProxy()
	proxyURL, err := url.Parse(cp.String())
	if err != nil {
		return fmt.Errorf("failed to parse proxy url: %w", err)
	}
	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return fmt.Errorf("failed to create dialer: %w", err)
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
			log.Warnf("failed to update dialer: %v", err)
			continue
		}
		r.makeHttpClient()
		r.advanceProxyIndex()
		log.Infof("current proxy: %v\n", r.currentProxy())
		wg.Done()
	}
}

func (r *Rotation) currentProxy() Proxy {
	return r.proxies[r.currentProxyIdx]
}

func (r *Rotation) advanceProxyIndex() {
	r.currentProxyIdx = (r.currentProxyIdx + 1) % len(r.proxies)
}
