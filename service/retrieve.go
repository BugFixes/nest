package service

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "os"

  "github.com/aws/aws-lambda-go/events"
  "github.com/bugfixes/agent"
)

// FindBug ....
func FindBug(request events.APIGatewayProxyRequest) (Response, error) {
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

  connection := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    os.Getenv("DB_HOSTNAME"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_USERNAME"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_DATABASE"),
  )

  searchKey := ""
  query := ""
  for param, val := range request.QueryStringParameters {
    switch param {
    case "hash":
      searchKey = val
      query = "SELECT agent_id, message, level, time_posted FROM bug WHERE agent_id = $1 AND hash = $2"
    case "bug":
      searchKey = val
      query = "SELECT agent_id, message, level, time_posted FROM bug WHERE agent_id = $1 AND id = $2"
    }
  }

  if query == "" {
    return Response{}, fmt.Errorf("FindBug no query: %+v", request)
  }

  bug := Bug{}

  db, err := sql.Open("postgres", connection)
  if err != nil {
    return Response{}, fmt.Errorf("FindBug sql.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("FindBug closedb: %+v", err)
    }
  }()
  row := db.QueryRow(query, agentID, searchKey)
  err = row.Scan(
    &bug.Agent,
    &bug.Message,
    &bug.Level,
    &bug.Time)
  if err != nil {
    return Response{}, fmt.Errorf("FindBug query: %w", err)
  }

  switch (bug.Level) {
  case 1:
    bug.LogLevel = "LOG"
  case 2:
    bug.LogLevel = "INFO"
  case 3:
    bug.LogLevel = "ERROR"
  default:
    bug.LogLevel = "OTHER"
  }

  resp, err := json.Marshal(bug)
  if err != nil {
    return Response{}, fmt.Errorf("FindBug marshall: %w", err)
  }

  return Response{
    Body: string(resp),
  }, nil
}
