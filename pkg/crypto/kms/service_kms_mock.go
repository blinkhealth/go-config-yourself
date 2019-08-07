// +build test

package kms

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	awsKMS "github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	log "github.com/sirupsen/logrus"
)

type mockKey struct {
	id        string
	errors    bool
	errorCode string
}

var BadCreds = awserr.New("NoCredentialProviders", "invalid-access-key", nil)
var mockKeys = []*mockKey{
	&mockKey{
		id:     "arn:aws:kms:us-east-1:000000000000:key/00000000-0000-0000-0000-000000000000",
		errors: false,
	},
	&mockKey{
		id:        "arn:aws:kms:us-east-1:111111111111:key/11111111-1111-1111-1111-111111111111",
		errors:    true,
		errorCode: "NoCredentialProviders",
	},
}

// newKMSClient creates a mock KMSClient
func newKMSClient(sess *session.Session, config *aws.Config) kmsiface.KMSAPI {
	log.Debug("Initializing mock KMS client")
	region := *config.Region
	creds, err := sess.Config.Credentials.Get()
	var accessKey = ""
	if err != nil {
		// Panic unless no credentials are set, which is a testable scenario
		if err, ok := err.(awserr.Error); !ok || err.Code() != "NoCredentialProviders" {
			panic(err)
		}
	} else {
		accessKey = creds.AccessKeyID
	}

	return &mockKMSClient{region: region, accessKey: accessKey}
}

type mockKMSClient struct {
	kmsiface.KMSAPI
	region    string
	accessKey string
}

type mockSecret struct {
	Key   string
	Value []byte
	Nonce string
}

func (m *mockKMSClient) Encrypt(input *awsKMS.EncryptInput) (*awsKMS.EncryptOutput, error) {
	if err := testKeyValidity(*input.KeyId); err != nil {
		return nil, err
	}
	if !testValidAccessKey(m.accessKey) {
		return nil, BadCreds
	}
	o := &awsKMS.EncryptOutput{}
	v, _ := json.Marshal(&mockSecret{
		Key:   *input.KeyId,
		Value: input.Plaintext,
		Nonce: strconv.FormatInt(time.Now().UnixNano(), 10),
	})

	o.SetCiphertextBlob(v)
	o.SetKeyId(*input.KeyId)
	return o, nil
}

func (m *mockKMSClient) Decrypt(input *awsKMS.DecryptInput) (*awsKMS.DecryptOutput, error) {
	if !testValidAccessKey(m.accessKey) {
		return nil, BadCreds
	}

	v := &mockSecret{}
	if err := json.Unmarshal(input.CiphertextBlob, &v); err != nil {
		return nil, errors.New("Wrong test secret stored!")
	}

	log.Debugf("Using key %s", v.Key)
	if err := testKeyValidity(v.Key); err != nil {
		return nil, err
	}

	o := &awsKMS.DecryptOutput{}
	o.SetPlaintext(v.Value)
	return o, nil
}

func (m *mockKMSClient) ListAliasesWithContext(context aws.Context, input *awsKMS.ListAliasesInput, opts ...request.Option) (*awsKMS.ListAliasesOutput, error) {
	if !testValidAccessKey(m.accessKey) {
		return nil, BadCreds
	}
	regionalKey := strings.Replace(mockKeys[0].id, "us-east-1", m.region, 1)
	aliases := []*awsKMS.AliasListEntry{
		&awsKMS.AliasListEntry{
			AliasArn: &regionalKey,
		},
	}

	return &awsKMS.ListAliasesOutput{
		NextMarker: nil,
		Truncated:  new(bool),
		Aliases:    aliases,
	}, nil
}

func testKeyValidity(keyId string) (err error) {
	if strings.Contains(keyId, ":000000000000:") {
		return nil
	}

	for _, key := range mockKeys {
		log.Debugf("testing %s %s", key.id, keyId)
		if key.id == keyId {
			if key.errors {
				return awserr.New(key.errorCode, key.errorCode, nil)
			} else {
				return nil
			}
		} else {
			continue
		}
	}

	msg := fmt.Sprintf("Could not find a key for %s", keyId)
	return awserr.New("UnknownTestKey", msg, nil)
}

func testValidAccessKey(accessKey string) bool {
	// log.Debugf("accessKey: %s")
	return accessKey == "AGOODACCESSKEYID"
}
