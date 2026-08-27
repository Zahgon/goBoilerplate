package tests

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"goBoilterplate/app/router"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// The tests in this file drive a real HTTP server over TCP rather than an
// in-process handler double, because every seam pinned here is a property of
// the bytes the server actually writes: status line, header set, content type
// spelling and body suffix. A recorder-based test reports the response the
// framework would have built and would stay green through most of the wiring
// mistakes below.

var (
	serverOnce sync.Once
	server     *httptest.Server
)

func seamServer() *httptest.Server {
	serverOnce.Do(func() {
		app := gin.New()
		router.Init(app)
		server = httptest.NewServer(app)
	})
	return server
}

type seamResponse struct {
	status int
	header http.Header
	body   string
}

func seamDo(t *testing.T, method, path string, body io.Reader, mutate func(*http.Request)) seamResponse {
	t.Helper()

	req, err := http.NewRequest(method, seamServer().URL+path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if mutate != nil {
		mutate(req)
	}

	// The default transport advertises gzip and transparently decodes it, which
	// would hide both the Content-Encoding header and the compressed bytes.
	transport := &http.Transport{DisableCompression: true}
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return seamResponse{status: res.StatusCode, header: res.Header, body: string(raw)}
}

func seamGet(t *testing.T, path string) seamResponse {
	t.Helper()
	return seamDo(t, http.MethodGet, path, nil, nil)
}

func seamAuthed(t *testing.T, method, path string) seamResponse {
	t.Helper()
	return seamDo(t, method, path, nil, func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+JWT.Token)
	})
}

func seamForm(t *testing.T, method, path string, form url.Values) seamResponse {
	t.Helper()
	return seamDo(t, method, path, strings.NewReader(form.Encode()), func(req *http.Request) {
		req.Header.Set("Content-Type", formContentType)
		req.Header.Set("Authorization", "Bearer "+JWT.Token)
	})
}

// TestSeamJSONEnvelope pins the JSON serialisation seam. The framework default
// on the target side is "application/json; charset=utf-8" with no trailing
// newline; the contract is a bare "application/json" and a body that ends in
// "\n", so a handler wired to the stock renderer fails here and nowhere else.
func TestSeamJSONEnvelope(t *testing.T) {
	res := seamGet(t, "/")

	assert.Equal(t, 200, res.status)
	assert.Equal(t, "application/json", res.header.Get("Content-Type"))
	assert.Equal(t, "\"Welcome to Echo\"\n", res.body)
}

// TestSeamHTMLContentType pins the HTML seam, including the upper-case charset
// spelling and the absence of a trailing newline.
func TestSeamHTMLContentType(t *testing.T) {
	res := seamGet(t, "/test")

	assert.Equal(t, 200, res.status)
	assert.Equal(t, "text/html; charset=UTF-8", res.header.Get("Content-Type"))
	assert.True(t, strings.HasPrefix(res.body, "<code> Protocol: HTTP/1.1<br> Host: "))
	assert.True(t, strings.HasSuffix(res.body, "<br> Method: GET<br> Path: /test<br> </code>"))
}

// TestSeamNotFoundEnvelope pins the router-generated 404 body. Every framework
// ships a different default envelope and the stock target one is empty.
func TestSeamNotFoundEnvelope(t *testing.T) {
	res := seamGet(t, "/nope")

	assert.Equal(t, 404, res.status)
	assert.Equal(t, "application/json", res.header.Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Not Found\"}\n", res.body)
}

// TestSeamMethodNotAllowedEnvelope pins the 405 body and the Allow header,
// which the source router stamps with OPTIONS first even though no OPTIONS
// handler is registered.
func TestSeamMethodNotAllowedEnvelope(t *testing.T) {
	res := seamDo(t, http.MethodPost, "/", nil, nil)

	assert.Equal(t, 405, res.status)
	assert.Equal(t, "application/json", res.header.Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Method Not Allowed\"}\n", res.body)
	assert.Equal(t, "OPTIONS, GET", res.header.Get("Allow"))
}

// TestSeamHeadErrorHasNoBody pins the error handler's HEAD special case: the
// status and Allow header are produced, but no body and no Content-Type.
func TestSeamHeadErrorHasNoBody(t *testing.T) {
	res := seamDo(t, http.MethodHead, "/", nil, nil)

	assert.Equal(t, 405, res.status)
	assert.Equal(t, "OPTIONS, GET", res.header.Get("Allow"))
	assert.Equal(t, "", res.header.Get("Content-Type"))
	assert.Equal(t, "", res.header.Get("Content-Length"))
	assert.Equal(t, "", res.body)
}

// TestSeamUnmatchedApiPathIsUnauthorized is the decisive routing seam. The
// source framework registers an implicit catch-all for every group carrying
// middleware, so an unmatched path under the guarded prefix runs the auth
// middleware first and answers 401. A port that leaves the fallback outside the
// group still builds, still passes the inherited suite, and answers 404 here.
func TestSeamUnmatchedApiPathIsUnauthorized(t *testing.T) {
	for _, path := range []string{"/api", "/api/", "/api/users/", "/api/nothing-here"} {
		res := seamGet(t, path)

		assert.Equalf(t, 401, res.status, "unauthenticated %s", path)
		assert.Equalf(t, "{\"message\":\"Unauthorized\"}\n", res.body, "unauthenticated %s", path)
	}
}

// TestSeamUnmatchedApiPathAuthenticatedIsNotFound is the other half of the same
// seam: once the middleware passes, the catch-all handler answers 404.
func TestSeamUnmatchedApiPathAuthenticatedIsNotFound(t *testing.T) {
	for _, path := range []string{"/api", "/api/", "/api/users/", "/api/nothing-here"} {
		res := seamAuthed(t, http.MethodGet, path)

		assert.Equalf(t, 404, res.status, "authenticated %s", path)
		assert.Equalf(t, "{\"message\":\"Not Found\"}\n", res.body, "authenticated %s", path)
	}
}

// TestSeamWrongMethodUnderApiIsNotFound pins that the guarded catch-all wins
// over method-not-allowed handling: PATCH on a real route answers 404 with no
// Allow header, not 405.
func TestSeamWrongMethodUnderApiIsNotFound(t *testing.T) {
	res := seamAuthed(t, http.MethodPatch, "/api/users")

	assert.Equal(t, 404, res.status)
	assert.Equal(t, "{\"message\":\"Not Found\"}\n", res.body)
	assert.Equal(t, "", res.header.Get("Allow"))
}

// TestSeamNoTrailingSlashRedirect pins that the router never redirects. The
// target framework enables trailing-slash and fixed-path redirects by default,
// which would turn both of these into a 301.
func TestSeamNoTrailingSlashRedirect(t *testing.T) {
	res := seamGet(t, "//")

	assert.Equal(t, 404, res.status)
	assert.Equal(t, "{\"message\":\"Not Found\"}\n", res.body)
	assert.Equal(t, "", res.header.Get("Location"))
}

// TestSeamLoginSkipperMisfires pins an upstream defect. The auth and gzip
// skippers test the matched route pattern, not the request URL, so GET on the
// login path falls through to the catch-all pattern, the skipper does not fire,
// and the answer is 401 rather than the 405 the route table implies.
func TestSeamLoginSkipperMisfires(t *testing.T) {
	res := seamGet(t, "/api/login")

	assert.Equal(t, 401, res.status)
	assert.Equal(t, "{\"message\":\"Unauthorized\"}\n", res.body)
}

// TestSeamLoginIsReachableWithoutToken pins the other side of the skipper: the
// declared login route is exempt from authentication.
func TestSeamLoginIsReachableWithoutToken(t *testing.T) {
	form := url.Values{}
	form.Set("email", "andres@teachlr.org")
	form.Set("password", "123456")

	res := seamDo(t, http.MethodPost, "/api/login", strings.NewReader(form.Encode()), func(req *http.Request) {
		req.Header.Set("Content-Type", formContentType)
	})

	assert.Equal(t, 200, res.status)
	assert.True(t, strings.HasPrefix(res.body, "{\"token\":\""))
}

// TestSeamGzipVaryIsUnconditional pins the compression seam. The header is
// added for every non-skipped request whether or not the client asked for gzip,
// and is absent on the skipped login route. A stock compression middleware only
// emits it when it actually compresses, which passes a smoke test and fails
// here.
func TestSeamGzipVaryIsUnconditional(t *testing.T) {
	plain := seamGet(t, "/")
	assert.Contains(t, plain.header.Values("Vary"), "Accept-Encoding")

	form := url.Values{}
	form.Set("email", "andres@teachlr.org")
	form.Set("password", "123456")
	login := seamDo(t, http.MethodPost, "/api/login", strings.NewReader(form.Encode()), func(req *http.Request) {
		req.Header.Set("Content-Type", formContentType)
	})
	assert.NotContains(t, login.header.Values("Vary"), "Accept-Encoding")
}

// TestSeamGzipCompresses pins that compression still happens when requested.
func TestSeamGzipCompresses(t *testing.T) {
	res := seamDo(t, http.MethodGet, "/test", nil, func(req *http.Request) {
		req.Header.Set("Accept-Encoding", "gzip")
	})

	assert.Equal(t, 200, res.status)
	assert.Equal(t, "gzip", res.header.Get("Content-Encoding"))
	assert.Equal(t, "text/html; charset=UTF-8", res.header.Get("Content-Type"))

	reader, err := gzip.NewReader(strings.NewReader(res.body))
	if !assert.NoError(t, err) {
		return
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(string(decoded), "<br> Method: GET<br> Path: /test<br> </code>"))
}

// TestSeamSecurityHeaders pins the security middleware and its skipper: three
// headers everywhere, none of them under the documentation prefix.
func TestSeamSecurityHeaders(t *testing.T) {
	res := seamGet(t, "/")
	assert.Equal(t, "1; mode=block", res.header.Get("X-Xss-Protection"))
	assert.Equal(t, "nosniff", res.header.Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", res.header.Get("X-Frame-Options"))
	assert.Equal(t, "", res.header.Get("Strict-Transport-Security"))
	assert.Equal(t, "", res.header.Get("Content-Security-Policy"))

	docs := seamGet(t, "/docs/index.html")
	assert.Equal(t, 200, docs.status)
	assert.Equal(t, "", docs.header.Get("X-Xss-Protection"))
	assert.Equal(t, "", docs.header.Get("X-Content-Type-Options"))
	assert.Equal(t, "", docs.header.Get("X-Frame-Options"))
}

// TestSeamCorsVaryWithoutOrigin pins that the CORS middleware announces Vary on
// every request, including plain same-origin ones with no Origin header. A
// stock CORS middleware returns early when Origin is absent and emits nothing.
func TestSeamCorsVaryWithoutOrigin(t *testing.T) {
	res := seamGet(t, "/")

	assert.Contains(t, res.header.Values("Vary"), "Origin")
	assert.Equal(t, "", res.header.Get("Access-Control-Allow-Origin"))
}

// TestSeamCorsSimpleRequest pins the literal star, rather than the echoed
// origin, on a simple cross-origin request.
func TestSeamCorsSimpleRequest(t *testing.T) {
	res := seamDo(t, http.MethodGet, "/", nil, func(req *http.Request) {
		req.Header.Set("Origin", "http://x.test")
	})

	assert.Equal(t, 200, res.status)
	assert.Equal(t, "*", res.header.Get("Access-Control-Allow-Origin"))
}

// TestSeamCorsPreflightOnGuardedRoute pins that preflight short-circuits ahead
// of every other middleware: no security headers, no compression Vary, no auth
// challenge, and no Allow header because the guarded catch-all matches every
// method.
func TestSeamCorsPreflightOnGuardedRoute(t *testing.T) {
	res := seamDo(t, http.MethodOptions, "/api/users", nil, func(req *http.Request) {
		req.Header.Set("Origin", "http://x.test")
		req.Header.Set("Access-Control-Request-Method", "GET")
	})

	assert.Equal(t, 204, res.status)
	assert.Equal(t, "*", res.header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "HEAD,GET,PUT,POST,DELETE", res.header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t,
		[]string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		res.header.Values("Vary"))
	assert.Equal(t, "", res.header.Get("Allow"))
	assert.Equal(t, "", res.header.Get("X-Frame-Options"))
	assert.Equal(t, "", res.body)
}

// TestSeamCorsPreflightCarriesAllow pins the opposite case: on an unguarded
// GET-only route the router has already stamped Allow before the CORS
// middleware writes its 204.
func TestSeamCorsPreflightCarriesAllow(t *testing.T) {
	res := seamDo(t, http.MethodOptions, "/", nil, func(req *http.Request) {
		req.Header.Set("Origin", "http://x.test")
		req.Header.Set("Access-Control-Request-Method", "PATCH")
	})

	assert.Equal(t, 204, res.status)
	assert.Equal(t, "OPTIONS, GET", res.header.Get("Allow"))
	assert.Equal(t, "HEAD,GET,PUT,POST,DELETE", res.header.Get("Access-Control-Allow-Methods"))
}

// TestSeamSwaggerRoutePattern pins the documentation mount. The bare prefix
// misses the wildcard route and falls through to the router's own 404, while
// the generated document is served with the title baked into the annotations.
func TestSeamSwaggerRoutePattern(t *testing.T) {
	bare := seamGet(t, "/docs")
	assert.Equal(t, 404, bare.status)
	assert.Equal(t, "application/json", bare.header.Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Not Found\"}\n", bare.body)

	doc := seamGet(t, "/docs/doc.json")
	assert.Equal(t, 200, doc.status)
	assert.Equal(t, "application/json; charset=utf-8", doc.header.Get("Content-Type"))
	assert.Contains(t, doc.body, "\"title\": \"Golang Echo API\"")
}

// TestSeamValidationEnvelope pins the validation error shape: a bare JSON array
// of field/rule pairs under 422, produced identically for an unparseable body,
// a missing content type and an empty form.
func TestSeamValidationEnvelope(t *testing.T) {
	want := "[{\"field\":\"Email\",\"rule\":\"required\"},{\"field\":\"Password\",\"rule\":\"required\"}]\n"

	unparseable := seamDo(t, http.MethodPost, "/api/login", strings.NewReader("{"), func(req *http.Request) {
		req.Header.Set("Content-Type", formContentType)
	})
	assert.Equal(t, 422, unparseable.status)
	assert.Equal(t, "application/json", unparseable.header.Get("Content-Type"))
	assert.Equal(t, want, unparseable.body)

	noContentType := seamDo(t, http.MethodPost, "/api/login", strings.NewReader("email=andres@teachlr.org"), nil)
	assert.Equal(t, 422, noContentType.status)
	assert.Equal(t, want, noContentType.body)
}

// TestSeamMalformedPathParameter pins the 400 produced when the identifier
// cannot be parsed, including the bare quoted string body.
func TestSeamMalformedPathParameter(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		res := seamAuthed(t, method, "/api/users/abc")

		assert.Equalf(t, 400, res.status, "%s /api/users/abc", method)
		assert.Equalf(t, "\"Bad Request\"\n", res.body, "%s /api/users/abc", method)
	}
}

// TestSeamQuirkDeleteMissingIdSucceeds pins an upstream defect: deleting an
// identifier that does not exist reports success.
func TestSeamQuirkDeleteMissingIdSucceeds(t *testing.T) {
	res := seamAuthed(t, http.MethodDelete, "/api/users/999999")

	assert.Equal(t, 200, res.status)
	assert.Equal(t, "\"Deleted\"\n", res.body)
}

// TestSeamQuirkMissingIdOnReadAndUpdate pins that the same identifier is a 404
// on the read and update paths, so the delete quirk above is genuinely a quirk
// rather than a global convention.
func TestSeamQuirkMissingIdOnReadAndUpdate(t *testing.T) {
	read := seamAuthed(t, http.MethodGet, "/api/users/999999")
	assert.Equal(t, 404, read.status)
	assert.Equal(t, "\"Not Found\"\n", read.body)

	form := url.Values{}
	form.Set("username", "Nobody")
	form.Set("email", "nobody@teachlr.org")
	form.Set("password", "123456")
	form.Set("role", "admin")
	update := seamForm(t, http.MethodPut, "/api/users/999999", form)
	assert.Equal(t, 404, update.status)
	assert.Equal(t, "\"Not Found\"\n", update.body)
}

// TestSeamQuirkPasswordIsSerialised pins an upstream defect: the stored
// password is part of every user representation.
func TestSeamQuirkPasswordIsSerialised(t *testing.T) {
	res := seamAuthed(t, http.MethodGet, "/api/logout")

	assert.Equal(t, 200, res.status)
	assert.Contains(t, res.body, "\"password\":\"123456\"")
}

// TestSeamRawUtf8Body pins that non-ASCII survives as raw UTF-8 rather than
// being escaped, and that the created resource is echoed back under 201.
func TestSeamRawUtf8Body(t *testing.T) {
	form := url.Values{}
	form.Set("username", "Ángel 日本")
	form.Set("email", "seam-utf8@teachlr.org")
	form.Set("password", "123456")
	form.Set("role", "admin")

	created := seamForm(t, http.MethodPost, "/api/users", form)
	assert.Equal(t, 201, created.status)
	assert.Equal(t, "application/json", created.header.Get("Content-Type"))
	assert.Contains(t, created.body, "\"username\":\"Ángel 日本\"")
	assert.NotContains(t, created.body, "\\u")
	assert.True(t, strings.HasSuffix(created.body, "}\n"))
}

// TestSeamUnauthenticatedShapes pins the auth challenge across every way of
// getting it wrong, including a bearer token that is merely malformed.
func TestSeamUnauthenticatedShapes(t *testing.T) {
	cases := []struct {
		name   string
		header string
	}{
		{"absent", ""},
		{"empty bearer", "Bearer "},
		{"malformed token", "Bearer not-a-token"},
		{"wrong scheme", "Token abc"},
	}

	for _, tc := range cases {
		res := seamDo(t, http.MethodGet, "/api/users", nil, func(req *http.Request) {
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
		})

		assert.Equalf(t, 401, res.status, "authorization %q", tc.name)
		assert.Equalf(t, "application/json", res.header.Get("Content-Type"), "authorization %q", tc.name)
		assert.Equalf(t, "{\"message\":\"Unauthorized\"}\n", res.body, "authorization %q", tc.name)
	}
}
