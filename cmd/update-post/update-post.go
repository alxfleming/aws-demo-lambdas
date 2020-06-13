package main

import (
	"aws-demo-lambdas/internal/model"
	"aws-demo-lambdas/internal/util"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"os"
	"time"
)

type Expression struct {
	Message          string    `dynamodbav:"Message"`
	UpdatedTimestamp time.Time `dynamodbav:"UpdatedTimestamp"`
}

type Key struct {
	UserId           string    `dynamodbav:"UserId"`
	MessageId        string    `dynamodbav:"MessageId"`
	CreatedTimestamp time.Time `dynamodbav:"CreatedTimestamp"`
}

func handle(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	svc := util.InitDynamoConnection()
	username := util.GetUsername(req)

	var post model.Post
	err := json.Unmarshal([]byte(req.Body), &post)
	if err != nil {
		fmt.Println("Couldn't unmarshall request body: ")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if post.UserId != username {
		return events.APIGatewayProxyResponse{StatusCode: 403}, errors.New("calling user is trying to edit a post that does not belong to them")
	}

	expr, err := dynamodbattribute.MarshalMap(Expression{
		Message:          post.Message,
		UpdatedTimestamp: post.UpdatedTimestamp,
	})
	if err != nil {
		fmt.Println("Got error marshalling expression:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	key, err := dynamodbattribute.MarshalMap(Key{
		UserId:           post.UserId,
		MessageId:        post.MessageId,
		CreatedTimestamp: post.CreatedTimestamp,
	})
	if err != nil {
		fmt.Println("Got error marshalling key:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: expr,
		TableName:                 aws.String("posts"),
		Key:                       key,
	}

	_, err = svc.UpdateItem(input)
	if err != nil {
		fmt.Println("Couldn't update post:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	jsonOut, err := json.Marshal(post)
	if err != nil {
		fmt.Println("Couldn't marshall post for output:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonOut),
	}, nil
}

func main() {
	lambda.Start(handle)
}
