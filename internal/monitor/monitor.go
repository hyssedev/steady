package monitor

import "net/url"

type Monitor struct {
	ID   int64
	Name string
	URL  *url.URL
}
