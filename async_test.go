package async

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_httpRequest(t *testing.T) {
	server := buildServer(t, http.StatusOK, 0)
	defer server.Close()

	host := server.URL
	req, _ := http.NewRequest(http.MethodGet, host+"/uri", nil)
	req.Header.Add("Accept", "application/json")

	client := server.Client()
	client.Do(req)
}

func Test_AddHTTPRequest(t *testing.T) {
	server := buildServer(t, http.StatusOK, 0)
	defer server.Close()

	host := server.URL
	req, _ := http.NewRequest(http.MethodGet, host+"/uri", nil)
	req.Header.Add("Accept", "application/json")

	requests := make([]RequestWrap, 0)

	wrap := RequestWrap{request: req, client: server.Client()}

	requests = append(requests, wrap)

	responses := requestAll(requests)

	assert.Equal(t, 1, len(responses))

}

func Test_addMultiHTTPRequest(t *testing.T) {
	server := buildServer(t, http.StatusOK, 1*time.Second)
	defer server.Close()

	host := server.URL

	requests := make([]RequestWrap, 0)

	requests = append(requests, RequestWrap{request: buildRequest(host), client: server.Client()})
	requests = append(requests, RequestWrap{request: buildRequest(host), client: server.Client()})
	requests = append(requests, RequestWrap{request: buildRequest(host), client: server.Client()})

	before := time.Now()
	responses := requestAll(requests)
	after := time.Now().Sub(before)

	assert.Equal(t, 3, len(responses), "Got 3 responses")

	t.Logf("Took %d milliseconds", after.Milliseconds())

	for _, response := range responses {
		assert.Equal(t, response.response.StatusCode, 200)
	}

	assert.LessOrEqual(t, after.Microseconds(), 1100*time.Millisecond)

}

func Test_AddMultiError(t *testing.T) {
	server := buildServer(t, http.StatusOK, 0)
	fail_server := buildServer(t, http.StatusBadRequest, 0)
	defer server.Close()
	defer fail_server.Close()

	requests := make([]RequestWrap, 0)

	requests = append(requests, RequestWrap{request: buildRequest(server.URL), client: server.Client()})
	requests = append(requests, RequestWrap{request: buildRequest(server.URL), client: server.Client()})
	requests = append(requests, RequestWrap{request: buildRequest(fail_server.URL), client: fail_server.Client()})

	responses := requestAll(requests)

	assert.Equal(t, 3, len(responses), "Got 3 responses")

	fail_count := 0
	success_count := 0
	for _, response := range responses {
		if response.response.StatusCode != http.StatusOK {
			fail_count++
		} else {
			success_count++
		}
	}
	assert.Equal(t, 1, fail_count)

}

func buildServer(t *testing.T, code int, wait time.Duration) *httptest.Server {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/uri")
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		time.Sleep(wait)
		w.WriteHeader(code)
		w.Write([]byte(`{"response":"value"}`))
	}))
	return server
}

func buildRequest(host string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, host+"/uri", nil)
	req.Header.Add("Accept", "application/json")
	return req
}
