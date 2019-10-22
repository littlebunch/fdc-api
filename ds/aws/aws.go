// Package aws implements the DataStore interface for DynamoDB
package aws

import (
	"log"

	fdc "github.com/littlebunch/fdc-api/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Aws implements a DataSource interface to a DynamoDB instance
type Aws struct {
	Conn *dynamodb.DynamoDB
}

// ConnectDs get an AWS DynamoDB client connection.
// This assumes you have AWS credentials set-up see:https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/
func (ds *Aws) ConnectDs(cs fdc.Config) error {
	var err error
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cs.Aws.Region)},
	)
	if err != nil {
		log.Fatalln("Cannot connect to Dynamodb!", err)
	} else {
		ds.Conn = dynamodb.New(sess)
	}
	return err
}

// Get finds data for a single food
func (ds Aws) Get(q string, f interface{}) error {
	result, err := ds.Conn.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(aws.Config.Table),
		Key: map[string]*dynamodb.AttributeValue{
			"FdcId": {
				N: aws.String(q),
			},
		},
	})
	f = result.Item
	return err
}
