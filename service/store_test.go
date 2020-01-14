package service_test

import (
  "github.com/aws/aws-lambda-go/events"
  "github.com/bugfixes/nest/service"
  "github.com/joho/godotenv"
  "github.com/stretchr/testify/assert"
  "os"
  "testing"
)

func TestBug_Store(t *testing.T) {
  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      t.Errorf("TestBug_Store godotenv err: %w", err)
    }
  }

  tests := []struct {
    name string
    bug  service.Bug
    err  error
  }{
    {
      name: "tester info",
      bug: service.Bug{
        Agent:    "42e14f47-323f-40e6-883e-f552425a3983",
        LogLevel: "info",
        Message:  "tester",
        Hash:     "9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
      },
    },
  }

  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      b, err := test.bug.GenerateIdentifier()
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("generateident: %w", err)
      }

      err = b.Store()
      passed = assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("store test: %w", err)
      }
    })
  }
}

func TestCreateBug(t *testing.T) {
  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      t.Errorf("TestCreateBug godotenv: %w", err)
    }
  }

  tests := []struct {
    name     string
    request  events.APIGatewayProxyRequest
    response service.Response
    err      error
  }{
    {
      name: "create basic bug",
      request: events.APIGatewayProxyRequest{
        Resource:   "/bug",
        HTTPMethod: "POST",
        Headers: map[string]string{
          "x-agent-id": "42e14f47-323f-40e6-883e-f552425a3983",
        },
        Body: "tester",
      },
    },
  }

  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      resp, err := service.CreateBug(test.request)
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("CreateBug err type: %w, %+v", err, test.err)
      }
      passed = assert.Equal(t, test.err, err)
      if !passed {
        t.Errorf("CreateBug err equal: %w, %+v", err, test.err)
      }
      passed = assert.IsType(t, test.response, resp)
      if !passed {
        t.Errorf("CreateBug response: %+v, %+v", resp, test.response)
      }
    })
  }
}
