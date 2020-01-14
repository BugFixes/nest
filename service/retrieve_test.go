package service_test

import (
  "github.com/aws/aws-lambda-go/events"
  "github.com/bugfixes/nest/service"
  "github.com/joho/godotenv"
  "github.com/stretchr/testify/assert"
  "os"
  "testing"
)

func TestFindBug(t *testing.T) {
  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      t.Errorf("TestFindBug godotenv: %w", err)
    }
  }

  tests := []struct {
    name     string
    request  events.APIGatewayProxyRequest
    response service.Response
    err      error
  }{
    {
      name: "find simple id",
      request: events.APIGatewayProxyRequest{
        Resource:   "/bug",
        HTTPMethod: "GET",
        QueryStringParameters: map[string]string{
          "bug": "00a08d42-f97a-48e9-8574-29d0e76457dc",
        },
        Headers: map[string]string{
          "x-agent-id": "42e14f47-323f-40e6-883e-f552425a3983",
        },
      },
      response: service.Response{
        Body: `{"message":"tester","loglevel":"INFO","agent":"42e14f47-323f-40e6-883e-f552425a3983","level":2,"posted":"2020-01-11T23:10:38.915455Z"}`,
      },
    },
  }

  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      resp, err := service.FindBug(test.request)
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("FindBug err type:  %w, %+v", err, test.err)
      }
      passed = assert.Equal(t, test.err, err)
      if !passed {
        t.Errorf("FindBug err equal: %w, %+v", err, test.err)
      }
      passed = assert.Equal(t, test.response, resp)
      if !passed {
        t.Errorf("FindBug resp equal: %+v, %+v", resp, test.response)
      }
    })
  }
}
