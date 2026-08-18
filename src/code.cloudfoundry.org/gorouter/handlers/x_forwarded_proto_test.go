package handlers_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/gorouter/handlers"
)

var _ = Describe("X-Forwarded-Proto", func() {
	var (
		req        *http.Request
		res        *httptest.ResponseRecorder
		nextCalled bool
	)

	BeforeEach(func() {
		req, _ = http.NewRequest("GET", "/foo", nil)
		nextCalled = false
	})

	processAndGetUpdatedHeader := func(handler *handlers.XForwardedProto) string {
		recordedRequest := &http.Request{}
		mockNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recordedRequest = r
			nextCalled = true
		})
		res = httptest.NewRecorder()
		handler.ServeHTTP(res, req, mockNext)
		return recordedRequest.Header.Get("X-Forwarded-Proto")
	}

	Context("when the SkipSanitization is true", func() {
		var handler *handlers.XForwardedProto
		BeforeEach(func() {
			handler = &handlers.XForwardedProto{
				SkipSanitization:         func(req *http.Request) bool { return true },
				ForceForwardedProtoHttps: false,
				SanitizeForwardedProto:   false,
			}
		})

		It("only calls next handler", func() {
			processAndGetUpdatedHeader(handler)
			Expect(nextCalled).To(BeTrue())
		})
		// This is when request is back from route services and it should not be touched
		It("does not sanitize X-Forwarded-Proto", func() {
			req.Header.Set("X-Forwarded-Proto", "http")
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("http"))
		})

		It("doesn't overwrite X-Forwarded-Proto if present", func() {
			req.Header.Set("X-Forwarded-Proto", "https")
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("https"))
		})

		It("does not sanitize Forwarded", func() {
			req.Header.Set("Forwarded", "proto=https")
			processAndGetUpdatedHeader(handler)
			Expect(req.Header.Get("Forwarded")).To(Equal("proto=https"))
		})
	})

	Context("when the ForceForwardedProtoHttps is true", func() {
		var handler *handlers.XForwardedProto
		BeforeEach(func() {
			handler = &handlers.XForwardedProto{
				SkipSanitization:         func(req *http.Request) bool { return false },
				ForceForwardedProtoHttps: true,
				SanitizeForwardedProto:   false,
			}
		})

		It("overrides X-Forwarded-Proto if present", func() {
			req.Header.Set("X-Forwarded-Proto", "http")
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("https"))
			Expect(nextCalled).To(BeTrue())
		})

		It("sets X-Forwarded-Proto to https if not present", func() {
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("https"))
			Expect(nextCalled).To(BeTrue())
		})

		It("replaces a client-supplied Forwarded header", func() {
			req.RemoteAddr = "10.0.0.1:4711"
			req.Header.Set("Forwarded", "proto=http")
			processAndGetUpdatedHeader(handler)
			Expect(req.Header.Values("Forwarded")).To(Equal([]string{"for=10.0.0.1;proto=https"}))
		})
	})

	Context("when the SanitizeForwardedProto is true", func() {
		var handler *handlers.XForwardedProto
		BeforeEach(func() {
			handler = &handlers.XForwardedProto{
				SkipSanitization:         func(req *http.Request) bool { return false },
				ForceForwardedProtoHttps: false,
				SanitizeForwardedProto:   true,
			}
		})

		It("sets X-Forwarded-Proto to http when connecting over http with header set to https", func() {
			req.Header.Set("X-Forwarded-Proto", "https")
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("http"))
			Expect(nextCalled).To(BeTrue())
		})

		It("sets X-Forwarded-Proto to https when connecting over https with header set to http", func() {
			req.Header.Set("X-Forwarded-Proto", "http")
			req.TLS = &tls.ConnectionState{}
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("https"))
			Expect(nextCalled).To(BeTrue())
		})

		It("sets X-Forwarded-Proto to http if client is not providing one and connecting over http", func() {
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("http"))
			Expect(nextCalled).To(BeTrue())
		})

		It("replaces a Forwarded header claiming https on a cleartext connection", func() {
			req.RemoteAddr = "10.0.0.1:4711"
			req.Header.Set("Forwarded", "for=1.2.3.4;proto=https")
			processAndGetUpdatedHeader(handler)
			Expect(req.Header.Values("Forwarded")).To(Equal([]string{"for=10.0.0.1;proto=http"}))
		})

		It("replaces every value the client sent, not just the first", func() {
			req.RemoteAddr = "10.0.0.1:4711"
			req.Header.Add("Forwarded", "proto=https")
			req.Header.Add("Forwarded", "proto=https")
			processAndGetUpdatedHeader(handler)
			Expect(req.Header.Values("Forwarded")).To(Equal([]string{"for=10.0.0.1;proto=http"}))
		})

		DescribeTable("renders the peer address as an RFC 7239 node",
			func(remoteAddr, expected string) {
				req.RemoteAddr = remoteAddr
				req.Header.Set("Forwarded", "for=1.2.3.4")
				processAndGetUpdatedHeader(handler)
				Expect(req.Header.Get("Forwarded")).To(Equal(expected))
			},
			Entry("IPv4 is a token", "10.0.0.1:4711", "for=10.0.0.1;proto=http"),
			Entry("IPv6 is bracketed and quoted", "[2001:db8::1]:4711", `for="[2001:db8::1]";proto=http`),
			Entry("IPv4-mapped IPv6 is bracketed and quoted, since it is written with colons",
				"[::ffff:192.0.2.1]:4711", `for="[::ffff:192.0.2.1]";proto=http`),
			Entry("an IPv6 zone is not a valid nodename, so for is left out", "[fe80::1%eth0]:4711", "proto=http"),
			Entry("a hostname is not a valid nodename, so for is left out", "lb.example.com:4711", "proto=http"),
			Entry("an address that does not split leaves out for", "not-an-address", "proto=http"),
		)

		It("replaces a Forwarded header the client sent empty", func() {
			req.RemoteAddr = "10.0.0.1:4711"
			req.Header["Forwarded"] = []string{""}
			processAndGetUpdatedHeader(handler)
			Expect(req.Header.Values("Forwarded")).To(Equal([]string{"for=10.0.0.1;proto=http"}))
		})

		It("does not add a Forwarded header when the client did not send one", func() {
			processAndGetUpdatedHeader(handler)
			Expect(req.Header).NotTo(HaveKey("Forwarded"))
		})
	})

	Context("when the client does not provide an X-Forwarded-Proto header with every property to false", func() {
		var handler *handlers.XForwardedProto
		BeforeEach(func() {
			handler = &handlers.XForwardedProto{
				SkipSanitization:         func(req *http.Request) bool { return false },
				ForceForwardedProtoHttps: false,
				SanitizeForwardedProto:   false,
			}
		})

		It("sets X-Forwarded-Proto to http when connecting over http with header not set", func() {
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("http"))
			Expect(nextCalled).To(BeTrue())
		})

		It("sets X-Forwarded-Proto to https when connecting over https with header not set", func() {
			req.TLS = &tls.ConnectionState{}
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("https"))
			Expect(nextCalled).To(BeTrue())
		})

		It("sets X-Forwarded-Proto to http if client is not providing one and connecting over http", func() {
			Expect(processAndGetUpdatedHeader(handler)).To(Equal("http"))
			Expect(nextCalled).To(BeTrue())
		})

		It("passes a client-supplied Forwarded header through", func() {
			req.Header.Set("Forwarded", "for=1.2.3.4;proto=https")
			processAndGetUpdatedHeader(handler)
			Expect(req.Header.Get("Forwarded")).To(Equal("for=1.2.3.4;proto=https"))
		})
	})
})
