package service

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "os"
  "time"

  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/sqs"
  "github.com/bugfixes/agent"
)

// CreateBug ...
func CreateBug(request events.APIGatewayProxyRequest) (Response, error) {
  queueURL := fmt.Sprintf("%s/queue/%s", os.Getenv("SQS_ENDPOINT"), os.Getenv("SQS_QUEUE"))

  agentID, err := agent.ConnectDetails{
    Host:     os.Getenv("DB_HOSTNAME"),
    Port:     os.Getenv("DB_PORT"),
    Username: os.Getenv("DB_USERNAME"),
    Password: os.Getenv("DB_PASSWORD"),
    Database: os.Getenv("DB_DATABASE"),
    Full:     "",
  }.FindAgentFromHeaders(request.Headers)
  if err != nil {
    return Response{}, fmt.Errorf("createBug invalid agentid: %+v", request)
  }

  s, err := session.NewSession(&aws.Config{
    Region:   aws.String(os.Getenv("SQS_REGION")),
    Endpoint: aws.String(os.Getenv("SQS_ENDPOINT")),
  })
  if err != nil {
    return Response{}, fmt.Errorf("createBug session: %w", err)
  }
  svc := sqs.New(s)

  result, err := svc.SendMessage(&sqs.SendMessageInput{
    MessageAttributes: map[string]*sqs.MessageAttributeValue{
      "clientId": &sqs.MessageAttributeValue{
        DataType:    aws.String("String"),
        StringValue: aws.String(agentID),
      },
    },
    MessageBody: aws.String(request.Body),
    QueueUrl:    aws.String(queueURL),
  })
  if err != nil {
    return Response{}, fmt.Errorf("CreateBug sendMessage: %w", err)
  }

  return Response{
    Headers: map[string]string{
      "x-queue-id": *result.MessageId,
    },
    Body: "bug queued",
  }, nil
}

// FileBug add the bug to the system
func FileBug(agent, body string) (Bug, error) {
  b := Bug{}
  err := json.Unmarshal([]byte(body), &b)
  if err != nil {
    return b, fmt.Errorf("unmarshal: %w", err)
  }
  b.Level = convertLevelFromString(b.LogLevel)
  b.Time = time.Now()

  if agent == "" {
    return b, fmt.Errorf("FileBug: agent invalid")
  }
  b.Agent = agent

  b, err = b.GenerateHash()
  if err != nil {
    return b, fmt.Errorf("generateHash: %w", err)
  }

  b, err = b.GenerateIdentifier()
  if err != nil {
    return b, fmt.Errorf("generateidentifier: %w", err)
  }

  err = b.Store()
  if err != nil {
    return b, fmt.Errorf("store: %w", err)
  }

  return b, nil
}

// Store inject into db
func (b Bug) Store() error {
  if b.Hash == "" {
    return fmt.Errorf("no hash given")
  }

  b.Time = time.Now()

  db, err := sql.Open("postgres", fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    os.Getenv("DB_HOSTNAME"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_USERNAME"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_DATABASE")))
  if err != nil {
    return fmt.Errorf("nest.Store sqlOpen: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("nest.Store dbClose: %+v", err)
    }
  }()
  _, err = db.Exec(
    "INSERT INTO bug (id, hash, message, agent_id, level, time_posted) VALUES ($1, $2, $3, $4, $5, $6)",
    b.Identifier,
    b.Hash,
    b.Message,
    b.Agent,
    b.Level,
    b.Time)
  if err != nil {
    return fmt.Errorf("nest.Store insert: %w", err)
  }

  return nil
}
