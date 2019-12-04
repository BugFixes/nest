package service

import (
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
)

type Bug struct {
	Message  string `json:"message"`
	LogLevel string `json:"loglevel"`

	Agent string

	Level      LogLevel
	Hash       string
	Identifier string
	Time       time.Time
}

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelLog   LogLevel = "log"
	LogLevelError LogLevel = "error"
)

type LogLevel string

func convertLevelToString(l LogLevel) string {
	switch l {
	case LogLevelInfo:
		return "info"
	case LogLevelLog:
		return "log"
	}

	return "error"
}

func convertLevelFromString(s string) LogLevel {
	switch s {
	case "info":
		return LogLevelInfo
	case "log":
		return LogLevelLog
	}

	return LogLevelError
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.Path != "/bug" {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("unknown endpoint: %+v", request)
	}

	agent := ""
	for h, v := range request.Headers {
	    hl := strings.ToLower(h)
	    if hl == "x-api-id" {
	        agent = v
        }
    }
	if agent == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "",
		}, fmt.Errorf("no agent: %+v", request.Headers)
	}

	if len(request.Body) == 0 {
	    return events.APIGatewayProxyResponse{
	        StatusCode: 200,
	        Body: "",
        }, fmt.Errorf("no body: %+v", request)
    }

	err := FileBug(agent, request.Body)
	if err != nil {
		fmt.Printf("%v err: %v\n", request.Resource, err)
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "bug stored",
	}, nil
}

func FileBug(agent string, body string) error {
	b := Bug{}
	err := json.Unmarshal([]byte(body), &b)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	b.Level = convertLevelFromString(b.LogLevel)
	b.Time = time.Now()
	b.Agent = agent

	b, err = b.GenerateHash()
	if err != nil {
		return fmt.Errorf("generateHash: %w", err)
	}

	b, err = b.GenerateIdentifier()
	if err != nil {
		return fmt.Errorf("generateidentifier: %w", err)
	}

	err = b.Store()
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	return nil
}

func (b Bug) Store() error {
	if b.Hash == "" {
		return fmt.Errorf("no hash given")
	}

	s, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("DB_REGION")),
		Endpoint: aws.String(os.Getenv("DB_ENDPOINT")),
	})
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}

	svc := dynamodb.New(s)
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE")),
		Item: map[string]*dynamodb.AttributeValue{
			"identifier": {
				S: aws.String(b.Identifier),
			},
			"bug_hash": {
				S: aws.String(b.Hash),
			},
			"bug": {
				S: aws.String(b.Message),
			},
			"level": {
				S: aws.String(b.LogLevel),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(#IDENTIFIER)"),
		ExpressionAttributeNames: map[string]*string{
			"#IDENTIFIER": aws.String("identifier"),
		},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return fmt.Errorf("inputerr errcode: %w", err)
			case "ValidationException":
				return fmt.Errorf("inputerr validaton: %w", err)
			default:
				return fmt.Errorf("inputerr unknown: %w", err)
			}
		}
	}

	return nil
}

func (b Bug) GenerateIdentifier() (Bug, error) {
	pre := fmt.Sprintf("%s%d", b.Agent, b.Time.Unix())
	b.Identifier = fmt.Sprintf("%x", sha256.Sum256([]byte(pre)))

	return b, nil
}

func (b Bug) GenerateHash() (Bug, error) {
	b.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(b.Message)))

	return b, nil
}
