package service_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/bugfixes/nest/service"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func injectAgent(id string) error {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOSTNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE")))
	if err != nil {
		return fmt.Errorf("injectAgent sqlOpen: %w", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("injectAgent dbClose: %+v", err)
		}
	}()
	_, err = db.Exec("INSERT INTO agent (id, name, key, secret, company_id) VALUES ($1, 'tester', $1, $1, $1)", id)
	if err != nil {
		return fmt.Errorf("injectAgent insert: %w", err)
	}

	return nil
}

func deleteAgent(id string) error {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOSTNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE")))
	if err != nil {
		return fmt.Errorf("deleteAgent sqlOpen: %w", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("deleteAgent dbClose: %+v", err)
		}
	}()
	_, err = db.Exec("DELETE FROM agent WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleteAgent insert: %w", err)
	}

	return nil
}

func injectBug() error {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOSTNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE")))
	if err != nil {
		return fmt.Errorf("injectBug sqlOpen: %w", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("injectBug dbClose: %+v", err)
		}
	}()
	_, err = db.Exec("INSERT INTO bug (id, hash, message, agent_id, level, time_posted) VALUES ($1, $2, $3, $4, $5, $6)",
		"2894165a-5abe-46f5-b848-fc87e3f267e5",
		"9bba5c53a0545e0c80184b946153c9f58387e3bd1d4ee35740f29ac2e718b019",
		"tester",
		"42e14f47-323f-40e6-883e-f552425a3983",
		"2",
		"2020-01-18 01:31:27.135868")
	if err != nil {
		return fmt.Errorf("injectBug insert: %w", err)
	}

	return nil
}

func deleteBug() error {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOSTNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE")))
	if err != nil {
		return fmt.Errorf("deleteBug sqlOpen: %w", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("deleteBug dbClose: %+v", err)
		}
	}()
	_, err = db.Exec("DELETE FROM bug WHERE id = $1", "2894165a-5abe-46f5-b848-fc87e3f267e5")
	if err != nil {
		return fmt.Errorf("deleteBug insert: %w", err)
	}

	return nil
}

func TestFileBug(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("TestFileBug godotenv err: %w", err)
		}
	}

	tests := []struct {
		name  string
		agent string
		body  string
		err   error
	}{
		{
			name:  "file tester",
			agent: "42e14f47-323f-40e6-883e-f552425a3983",
			body:  `{"message":"tester","loglevel":"info"}`,
		},
	}

	for _, test := range tests {
		err := injectAgent(test.agent)
		if err != nil {
			t.Errorf("%v inject: %w", test.name, err)
		}

		t.Run(test.name, func(t *testing.T) {
			_, err := service.FileBug(test.agent, test.body)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("service test: %w", err)
			}
		})

		err = deleteAgent(test.agent)
		if err != nil {
			t.Errorf("%v delete: %w", test.name, err)
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
