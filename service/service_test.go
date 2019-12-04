package service_test

import (
	"testing"

	"github.com/bugfixes/nest/service"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestFileBug(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Errorf("godotenv: %w", err)
	}

	tests := []struct {
		name string
		body string
		err  error
	}{
		{
			name: "test",
			body: `{"message":"tester","loglevel":"info"}`,
		},
	}

	for _, test := range tests {
		err := service.FileBug(test.name, test.body)
		passed := assert.IsType(t, test.err, err)
		if !passed {
			t.Errorf("service test: %w", err)
		}
	}
}

func TestGenerateHash(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Errorf("godotenv: %w", err)
	}

	tests := []struct {
		bug    service.Bug
		expect string
		err    error
	}{
		{
			bug: service.Bug{
				Agent:    "tester",
				LogLevel: "info",
				Message:  "tester",
			},
			expect: "9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
		},
	}

	for _, test := range tests {
		resp, err := test.bug.GenerateHash()
		passed := assert.IsType(t, test.err, err)
		if !passed {
			t.Errorf("generatehash err: %w", err)
		}
		passed = assert.Equal(t, test.expect, resp.Hash)
		if !passed {
			t.Errorf("generatehash: %w", err)
		}
	}
}

func TestStoreBug(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Errorf("godotenv: %w", err)
	}

	tests := []struct {
		bug service.Bug
		err error
	}{
		{
			bug: service.Bug{
				Agent:    "tester",
				LogLevel: "info",
				Message:  "tester",
				Hash:     "9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
			},
		},
	}

	for _, test := range tests {
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
	}
}
