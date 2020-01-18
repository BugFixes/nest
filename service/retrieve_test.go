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
		agent    string
		request  events.APIGatewayProxyRequest
		response service.Response
		err      error
	}{
		{
			name:  "find simple id",
			agent: "42e14f47-323f-40e6-883e-f552425a3983",
			request: events.APIGatewayProxyRequest{
				Resource:   "/bug",
				HTTPMethod: "GET",
				QueryStringParameters: map[string]string{
					"bug": "2894165a-5abe-46f5-b848-fc87e3f267e5",
				},
				Headers: map[string]string{
					"x-agent-id": "42e14f47-323f-40e6-883e-f552425a3983",
				},
			},
			response: service.Response{
				Body: `{"message":"tester","loglevel":"INFO","agent":"42e14f47-323f-40e6-883e-f552425a3983","level":2,"posted":"2020-01-18T01:31:27.135868Z"}`,
			},
		},
	}

	for _, test := range tests {
		err := injectAgent(test.agent)
		if err != nil {
			t.Errorf("%v inject agent: %w", test.name, err)
		}

		err = injectBug()
		if err != nil {
			t.Errorf("%v inject bug: %w", test.name, err)
		}

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

		err = deleteAgent(test.agent)
		if err != nil {
			t.Errorf("%v delete agent: %w", test.name, err)
		}

		err = deleteBug()
		if err != nil {
			t.Errorf("%v delete bug: %w", test.name, err)
		}
	}
}
