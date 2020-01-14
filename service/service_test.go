package service_test

import (
  "os"
  "testing"

  "github.com/bugfixes/nest/service"
  "github.com/joho/godotenv"
  "github.com/stretchr/testify/assert"
)

func TestFileBug(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("TestFileBug godotenv err: %w", err)
		}
	}

	tests := []struct {
		agent string
		body  string
		err   error
	}{
		{
			agent: "42e14f47-323f-40e6-883e-f552425a3983",
			body:  `{"message":"tester","loglevel":"info"}`,
		},
	}

	for _, test := range tests {
		_, err := service.FileBug(test.agent, test.body)
		passed := assert.IsType(t, test.err, err)
		if !passed {
			t.Errorf("service test: %w", err)
		}
	}
}

func TestGenerateHash(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("TestGenerateHash godotenv err: %w", err)
		}
	}

	tests := []struct {
		name   string
		bug    service.Bug
		expect string
		err    error
	}{
		{
			name: "hash info tester",
			bug: service.Bug{
				Agent:    "42e14f47-323f-40e6-883e-f552425a3983",
				LogLevel: "info",
				Message:  "tester",
			},
			expect: "9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.bug.GenerateHash()
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("generatehash err: %w", err)
			}
			passed = assert.Equal(t, test.expect, resp.Hash)
			if !passed {
				t.Errorf("generatehash: %w", err)
			}
		})
	}
}
