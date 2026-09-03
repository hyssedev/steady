package monitor

import "net/url"

type Monitor struct {
	Name string
	URL  *url.URL
}
