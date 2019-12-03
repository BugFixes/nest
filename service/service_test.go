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

	tests := []struct{
		name string
		body string
		err error
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
