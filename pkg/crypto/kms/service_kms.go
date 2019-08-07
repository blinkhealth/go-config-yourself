// +build !test

package kms

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	awsKMS "github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

// newKMSClient returns the real KMSClient
func newKMSClient(sess client.ConfigProvider, config *aws.Config) kmsiface.KMSAPI {
	return awsKMS.New(sess, config)
}
