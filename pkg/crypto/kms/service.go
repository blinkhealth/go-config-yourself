package kms

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	awsKMS "github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	log "github.com/sirupsen/logrus"
)

// enough to be less annoying around network outages and give the user
// time to enter MFA
const keyListTimeout = 15 * time.Second

// This can't be a constant cause then we can't get a pointer to it :/
// https://golang.org/ref/spec#Address_operators
var keyListQueryLimit = int64(100)

type kmsService struct {
	session *session.Session
	config  *aws.Config
	client  kmsiface.KMSAPI
}

func stdinTokenProvider() (string, error) {
	var v string
	fmt.Fprintf(os.Stderr, "Enter an MFA token code for profile %s: ", os.Getenv("AWS_PROFILE"))
	_, err := fmt.Scanln(&v)

	return v, err
}

func createAWSSession(region string) (sess *session.Session, config *aws.Config) {
	config = aws.NewConfig().WithRegion(region)
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		Config:                  *config,
		AssumeRoleTokenProvider: stdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	}))
	return
}

func newKMSService(region string) (svc *kmsService) {
	awsSess, awsConfig := createAWSSession(region)

	return &kmsService{
		session: awsSess,
		config:  awsConfig,
		client:  newKMSClient(awsSess, awsConfig),
	}
}

// Encrypt a string with a kms key
func (svc *kmsService) Encrypt(key string, plainText []byte) ([]byte, error) {
	result, err := svc.client.Encrypt(&awsKMS.EncryptInput{
		KeyId:     &key,
		Plaintext: plainText,
	})

	if err != nil {
		return nil, catchBadCredentials(err, svc.session, key)
	}

	return result.CiphertextBlob, nil
}

// Decrypt a some bytes
func (svc *kmsService) Decrypt(encryptedBytes []byte) (string, error) {
	out, err := svc.client.Decrypt(&awsKMS.DecryptInput{
		CiphertextBlob: encryptedBytes,
	})
	if err != nil {
		return "", catchBadCredentials(err, svc.session, "")
	}

	return string(out.Plaintext), err
}

// ListKeys offers a list of kms keys on all regions
func (svc *kmsService) ListKeys() (keys []string, err error) {
	finished := make(chan bool, 1)
	listingErrors := make(chan error, 1)
	var wg sync.WaitGroup

	// Query every region for their known kms keys
	for _, region := range listRegions() {
		if region == "ap-east-1" && os.Getenv("AWS_AP_EAST_1_ENABLED") == "" {
			// https://docs.aws.amazon.com/general/latest/gr/rande.html
			continue
		}
		region := region
		wg.Add(1)
		go func(region string, listingErrors chan<- error) {
			regionConfig := svc.config.Copy(&aws.Config{Region: &region})
			client := newKMSClient(svc.session, regionConfig)
			defer wg.Done()
			log.Debugf("Querying for keys in %s", region)
			regionKeys, err := fetchAllKeys(client, nil)
			if err = catchBadCredentials(err, svc.session, region); err != nil {
				listingErrors <- fmt.Errorf("Error querying for keys in %s: %s", region, err)
				return
			}

			log.Debugf("Found %d customer keys on %s", len(regionKeys), region)

			if len(regionKeys) > 0 {
				keys = append(keys, regionKeys...)
			}
		}(region, listingErrors)
	}

	go func() {
		// Wait for all of querying to be done
		wg.Wait()
		close(finished)
		close(listingErrors)
	}()

	select {
	case <-finished:
	case err = <-listingErrors:
		if err != nil {
			return keys, err
		}
	}

	if len(keys) < 1 {
		err = errors.New("Could not find any KMS keys")
		return
	}

	sort.Strings(keys)
	return
}

func fetchAllKeys(client kmsiface.KMSAPI, nextMarker *string) (result []string, err error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), keyListTimeout)
	defer cancelFn()
	resp, err := client.ListAliasesWithContext(ctx, &awsKMS.ListAliasesInput{
		Limit:  &keyListQueryLimit,
		Marker: nextMarker,
	})

	if err != nil {
		return nil, err
	}

	customerKeys := make([]string, 0)
	for _, key := range resp.Aliases {
		// filters out kms keys created by AWS
		if strings.Contains(*key.AliasArn, "alias/aws") {
			continue
		}
		customerKeys = append(customerKeys, *key.AliasArn)
	}

	if *resp.Truncated {
		nextSet, err := fetchAllKeys(client, resp.NextMarker)
		return append(customerKeys, nextSet...), err
	}

	return customerKeys, nil
}

func listRegions() (regions []string) {
	partition := endpoints.AwsPartition()
	for regionName := range partition.Regions() {
		regions = append(regions, regionName)
	}

	return
}

func catchBadCredentials(err error, sess *session.Session, key string) error {
	if awsErr, ok := err.(awserr.Error); ok {
		code := awsErr.Code()
		switch code {
		case "AccessDeniedException":
			msg := fmt.Sprintf("AWS denied access in region <%s>", *sess.Config.Region)
			if key != "" {
				msg = fmt.Sprintf("%s using key <%s>", msg, key)
			}

			stsClient := sts.New(sess)
			result, stsErr := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			if stsErr == nil {
				msg = fmt.Sprintf("%s for identity <%s>", msg, *result.Arn)
			}

			return fmt.Errorf("%s (%s)", msg, code)
		case "NoCredentialProviders":
			return errors.New("No AWS credentials found")
		case "RequestCanceled":
			log.Warnf("Timed out before being able to fetch keys from region <%s>", key)
			return nil
		}
	}

	return err
}
