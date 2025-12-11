package feature

import (
	"fmt"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type HttpTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, &HttpTestSuite{})
}

func (s *HttpTestSuite) SetupSuite() {
}

// SetupTest will run before each test in the suite.
func (s *HttpTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *HttpTestSuite) TearDownTest() {
}


func (s *HttpTestSuite) TestBindQuery() {
	resp, err := s.Http(s.T()).Get("/bind-query?name=Goravel")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"name\":\"Goravel\"}", content)
}

func (s *HttpTestSuite) TestFallback() {
	resp, err := s.Http(s.T()).Get("/lang")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).Get("/not-found")
	s.Require().NoError(err)
	resp.AssertNotFound()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("fallback", content)
}

func (s *HttpTestSuite) TestFiles() {
	body, err := http.NewBody().SetFiles(map[string][]string{
		"files": []string{"lang/cn.json", "lang/en.json"},
	}).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Post("/files", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"files\":[\"cn.json\",\"en.json\"]}", content)
}

func (s *HttpTestSuite) TestInputMap() {
	body, err := http.NewBody().SetField("test", map[string]any{"key1": "value1", "key2": "value2"}).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Post("/input-map", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"test\":{\"key1\":\"value1\",\"key2\":\"value2\"}}", content)
}

func (s *HttpTestSuite) TestInputMapArray() {
	body, err := http.NewBody().SetField("test", []map[string]any{{"key1": "value1", "key2": "value2"}, {"key3": "value3", "key4": "value4"}}).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Post("/input-map-array", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"test\":[{\"key1\":\"value1\",\"key2\":\"value2\"},{\"key3\":\"value3\",\"key4\":\"value4\"}]}", content)
}

func (s *HttpTestSuite) TestLang() {
	tests := []struct {
		name           string
		lang           string
		expectResponse map[string]any
	}{
		{
			name:           "use default lang",
			expectResponse: map[string]any{"current_locale": "en", "fallback": "Goravel 是一个基于 Go 语言的 Web 开发框架", "name": "Goravel Framework"},
		},
		{
			name:           "lang is cn",
			lang:           "cn",
			expectResponse: map[string]any{"current_locale": "cn", "fallback": "Goravel 是一个基于 Go 语言的 Web 开发框架", "name": "Goravel 框架"},
		},
		{
			name:           "lang is fs",
			lang:           "fs",
			expectResponse: map[string]any{"current_locale": "fs", "fallback": "Goravel 是一个基于 Go 语言的 Web 开发框架", "name": "fs name"},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			resp, err := s.Http(s.T()).Get(fmt.Sprintf("/lang?lang=%s", test.lang))

			s.NoError(err)
			resp.AssertSuccessful()
			resp.AssertJson(test.expectResponse)
		})
	}
}

func (s *HttpTestSuite) TestPanic() {
	resp, err := s.Http(s.T()).Get("/panic")

	s.Require().NoError(err)
	resp.AssertInternalServerError()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("recover", content)
}

func (s *HttpTestSuite) TestStream() {
	resp, err := s.Http(s.T()).Get("/stream")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\n", content)
}

func (s *HttpTestSuite) TestThrottle() {
	resp, err := s.Http(s.T()).Get("/throttle")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).Get("/throttle")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).Get("/throttle")
	s.Require().NoError(err)
	resp.AssertTooManyRequests()
}

func (s *HttpTestSuite) TestTimeout() {
	resp, err := s.Http(s.T()).Get("/timeout")

	s.Require().NoError(err)
	resp.AssertStatus(contractshttp.StatusRequestTimeout)
}

func (s *HttpTestSuite) TestUrl() {
	resp, err := s.Http(s.T()).Get("/url/get/1?a=1&b=2")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal(`{"full_url":"http://example.com/url/get/1?a=1\u0026b=2","info":{"handler":"goravel/routes.Api.func11.1","method":"GET","name":"url.get","path":"/url/get/{id}"},"info1":{"handler":"goravel/routes.Api.func11.1","method":"GET|HEAD","name":"url.get","path":"/url/get/{id}"},"method":"GET","name":"url.get","origin_path":"/url/get/{id}","path":"/url/get/1","url":"/url/get/1?a=1\u0026b=2"}`, content)

	resp, err = s.Http(s.T()).Post("/url/post/1?a=1&b=2", strings.NewReader("{\"name\":\"Goravel\"}"))
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err = resp.Content()
	s.Require().NoError(err)
	s.Equal(`{"full_url":"http://example.com/url/post/1?a=1\u0026b=2","info":{"handler":"goravel/routes.Api.func11.2","method":"POST","name":"url.post","path":"/url/post/{id}"},"info1":{"handler":"goravel/routes.Api.func11.2","method":"POST","name":"url.post","path":"/url/post/{id}"},"method":"POST","name":"url.post","origin_path":"/url/post/{id}","path":"/url/post/1","url":"/url/post/1?a=1\u0026b=2"}`, content)
}

