package handlers

import (
	"net"
	"net/http"
	"strings"
)

type XForwardedProto struct {
	SkipSanitization         func(req *http.Request) bool
	ForceForwardedProtoHttps bool
	SanitizeForwardedProto   bool
}

func (h *XForwardedProto) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	newReq := new(http.Request)
	*newReq = *r
	skip := h.SkipSanitization(r)
	if !skip {
		if h.ForceForwardedProtoHttps {
			newReq.Header.Set("X-Forwarded-Proto", "https")
		} else if h.SanitizeForwardedProto || newReq.Header.Get("X-Forwarded-Proto") == "" {
			scheme := "http"
			if newReq.TLS != nil {
				scheme = "https"
			}
			newReq.Header.Set("X-Forwarded-Proto", scheme)
		}
		// RFC 7239 carries the same claim in its proto parameter, and backends
		// that read Forwarded tend to prefer it over X-Forwarded-Proto, so it
		// needs sanitizing too. This middleware has no trusted source for the
		// parameters the client sent, so replace the entire field with the
		// values gorouter can establish for this hop. Test for presence, not
		// content: an empty field value is still a field the client set.
		if (h.ForceForwardedProtoHttps || h.SanitizeForwardedProto) && len(newReq.Header.Values("Forwarded")) > 0 {
			newReq.Header.Set("Forwarded", forwardedElement(newReq))
		}
	}

	next(rw, newReq)
}

// forwardedElement renders the RFC 7239 element gorouter can vouch for: the
// immediate peer of the connection and the scheme sanitization settled on.
// The peer is whoever connected to gorouter, which behind a load balancer is
// the load balancer rather than the end client, so it is the hop gorouter can
// establish rather than necessarily the original client. It is the same
// immediate-peer address added to X-Forwarded-For. An app that gives Forwarded
// precedence will not consult X-Forwarded-For, so leaving out for would
// downgrade the address it sees to gorouter's own.
//
// host is left out on purpose. An app that finds no host parameter can fall
// back to the Host header, which should be the value gorouter routed on.
func forwardedElement(r *http.Request) string {
	element := "proto=" + r.Header.Get("X-Forwarded-Proto")

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return element
	}
	// Only an IP literal is a valid nodename. Anything else, a hostname or an
	// address carrying an IPv6 zone, is dropped rather than rendered invalid.
	if net.ParseIP(host) == nil {
		return element
	}
	// Classify on the text, not on the parsed address: an IPv4-mapped address
	// such as ::ffff:192.0.2.1 parses as IPv4 but is still written with colons,
	// which a token cannot hold. Colons mean bracket and quote.
	if strings.Contains(host, ":") {
		return `for="[` + host + `]";` + element
	}
	return "for=" + host + ";" + element
}
