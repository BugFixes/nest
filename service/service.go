package service

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/agent"
	_ "github.com/lib/pq" // drivers are usually blank
	uuid "github.com/satori/go.uuid"
)

// Bug ...
type Bug struct {
	Message  string `json:"message,omitempty"`
	LogLevel string `json:"loglevel,omitempty"`

	Agent string `json:"agent,omitempty"`

	AgentKey    string `json:"key,omitempty"`
	AgentSecret string `json:"secret,omitempty"`

	Level      int       `json:"level,omitempty"`
	Hash       string    `json:"hash,omitempty"`
	Identifier string    `json:"id,omitempty"`
	Time       time.Time `json:"posted,omitempty"`
}

func convertLevelFromString(s string) int {
	switch s {
	case "log":
		return 1
	case "info":
		return 2
	case "error":
		return 3
	default:
		return 4
	}
}

// Response ...
type Response struct {
	Body    string
	Headers map[string]string
}

// Handler ...
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.Path {
	case "/bug":
		response, err := processRequest(request)
		if err != nil {
			return events.APIGatewayProxyResponse{}, fmt.Errorf("Handler /bug: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil
	case "/file":
		response, err := processFile(request)
		if err != nil {
			return events.APIGatewayProxyResponse{}, fmt.Errorf("Handler /file: %w", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    response.Headers,
			Body:       response.Body,
		}, nil
	default:
		return events.APIGatewayProxyResponse{}, fmt.Errorf("unknown endpoint: %+v", request)
	}
}

func processFile(request events.APIGatewayProxyRequest) (Response, error) {
	agentID, err := agent.ConnectDetails{
		Host:     os.Getenv("DB_HOSTNAME"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
		Full:     "",
	}.FindAgentFromHeaders(request.Headers)
	if err != nil {
		return Response{}, fmt.Errorf("invalid agentid: %+v", request)
	}

	bug, err := FileBug(agentID, request.Body)
	if err != nil {
		fmt.Printf("%+v err: %+v\n", request.Resource, err)
		return Response{}, fmt.Errorf("processFile file: %w", err)
	}

	return Response{
		Headers: map[string]string{
			"x-bug-id": bug.Identifier,
		},
		Body: "bug filed",
	}, nil
}

func processRequest(request events.APIGatewayProxyRequest) (Response, error) {
	if len(request.Body) == 0 {
		return Response{}, fmt.Errorf("processRequest: no body: %+v", request)
	}

	switch request.HTTPMethod {
	case "POST":
		resp, err := CreateBug(request)
		if err != nil {
			return Response{}, fmt.Errorf("processRequest create: %w", err)
		}
		return resp, nil
	case "GET":
		resp, err := FindBug(request)
		if err != nil {
			return Response{}, fmt.Errorf("processRequest find: %w", err)
		}
		return resp, nil
	}

	return Response{}, nil
}

// GenerateIdentifier ...
func (b Bug) GenerateIdentifier() (Bug, error) {
	b.Identifier = uuid.NewV4().String()

	return b, nil
}

// GenerateHash ...
func (b Bug) GenerateHash() (Bug, error) {
	b.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(b.Message)))

	return b, nil
}
