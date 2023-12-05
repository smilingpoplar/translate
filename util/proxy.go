package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

func SetProxy(addr string, client *http.Client) error {
	if addr == "" {
		return nil
	}

	var transport *http.Transport
	if strings.HasPrefix(addr, "http") {
		proxyURL, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("error creating proxy %s: %w", addr, err)
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	} else if strings.HasPrefix(addr, "socks5://") {
		addrNoPrefix := strings.TrimPrefix(addr, "socks5://")
		dialer, err := proxy.SOCKS5("tcp", addrNoPrefix, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("error creating proxy %s: %w", addr, err)
		}
		transport = &http.Transport{
			Dial: dialer.Dial,
		}
	} else {
		return fmt.Errorf("unknown proxy %s", addr)
	}

	client.Transport = transport
	return nil
}
