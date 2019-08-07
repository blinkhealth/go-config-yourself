package gpg

import (
	"bytes"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/blinkhealth/go-config-yourself/internal/datakey"
	"github.com/proglottis/gpgme"
)

type gpgService struct {
	// The encrypted key
	encryptedKey *[]byte
	// The datakey Service
	dataKey *datakey.Service
	// The recipients for this key
	recipients []string
}

// Create a new GPG service from a password hash
func newGPGService(recipients []string) (svc *gpgService, err error) {
	var encryptedKey bytes.Buffer

	// Create a new key
	fileKey, err := datakey.New()
	if err != nil {
		return nil, err
	}

	keys, err := findKeys(recipients)
	if err != nil {
		return nil, err
	}

	plain, err := gpgme.NewDataBytes(fileKey)
	if err != nil {
		return nil, err
	}

	cipher, err := gpgme.NewDataReadWriter(&encryptedKey)
	if err != nil {
		return nil, err
	}

	ctx, err := gpgme.New()
	if err != nil {
		return nil, err
	}

	ctx.SetArmor(true)
	if err = ctx.Encrypt(keys, gpgme.EncryptAlwaysTrust, plain, cipher); err != nil {
		log.Debugf("Key generation encrypt failed: %s, %s", err, recipients)
		return
	}

	keyBytes := encryptedKey.Bytes()

	// Create a service with the encrypted key
	svc = &gpgService{
		encryptedKey: &keyBytes,
		dataKey:      datakey.NewService(fileKey),
		recipients:   recipients,
	}

	return svc, nil
}

// Hydrate a service from a persisted key
func gpgServiceFromConfig(encryptedKey string, recipients []string) (svc *gpgService) {
	key := []byte(encryptedKey)
	svc = &gpgService{
		encryptedKey: &key,
		recipients:   recipients,
	}

	return
}

// Serialize renders the service key
func (svc *gpgService) Serialize(data map[string]interface{}) {
	data["key"] = string(*svc.encryptedKey)
	data["recipients"] = svc.recipients
}

// IsAvailable tells if the service has a decrypted key
func (svc *gpgService) IsAvailable() bool {
	return svc.dataKey != nil
}

// DecryptKey decrypts the encrypted key with a given password
func (svc *gpgService) DecryptKey() (err error) {
	if svc.IsAvailable() {
		return
	}
	log.Debugf("Decrypting key")

	if svc.encryptedKey == nil {
		return fmt.Errorf("No key found")
	}

	keyData, err := gpgme.NewDataBytes(*svc.encryptedKey)
	if err != nil {
		return
	}

	plainKey, err := gpgme.Decrypt(keyData)
	if err != nil {
		return err
	}

	keyBuffer := new(bytes.Buffer)
	if _, err = io.Copy(keyBuffer, plainKey); err != nil {
		return
	}

	svc.dataKey = datakey.NewService(keyBuffer.Bytes())

	return
}

// Decrypt plaintext
func (svc *gpgService) Decrypt(encryptedBytes []byte) (plainText string, err error) {
	return svc.dataKey.Decrypt(encryptedBytes)
}

// Encrypt plaintext
func (svc *gpgService) Encrypt(plainText []byte) (cipherText []byte, err error) {
	return svc.dataKey.Encrypt(plainText)
}

// ListKeys lists all public keys
func (svc *gpgService) ListKeys() (keys []string, err error) {
	rcpt, err := gpgme.FindKeys("", false)
	if err != nil {
		log.Debugf("listing error: %s", err)
		return nil, fmt.Errorf("Unable to list all keys: %s", err)
	}

	for _, r := range rcpt {
		keys = append(keys, fmt.Sprintf("%v", r.UserIDs().Email()))
	}

	return
}

func findKeys(filters []string) (keys []*gpgme.Key, err error) {
	for _, filter := range filters {
		rcpt, err := gpgme.FindKeys(filter, false)
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch key for %s: %s", filter, err)
		}

		if len(rcpt) == 0 {
			return nil, fmt.Errorf("Could not find a key for %s", filter)
		}

		keys = append(keys, rcpt...)
	}

	return
}
