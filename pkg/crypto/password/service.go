package password

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/scrypt"

	"github.com/blinkhealth/go-config-yourself/internal/datakey"
)

const (
	saltSize = 12
	// From https://godoc.org/golang.org/x/crypto/scrypt
	// > N is a CPU/memory cost parameter, which must be a power of two greater than 1
	scryptCost = 32 * 1024
	// > r and p must satisfy r * p < 2^30
	scryptR = 8
	scryptP = 1
	// quote continues:
	// > The recommended parameters for interactive logins as of 2017 are N=32768, r=8 and p=1.
	// > The parameters N, r, and p should be increased as memory latency and CPU parallelism increases;
	// > consider setting N to the highest power of 2 you can derive within 100 milliseconds.
)

type passwordService struct {
	// The encrypted key
	encryptedKey *[]byte
	// The decrypted key
	dataKey *datakey.Service
	// The salt used to hash this password
	salt *[]byte
}

func keyFromPassword(password []byte, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, scryptCost, scryptR, scryptP, 32)
}

// Create a new password service from a password in plain-text
func newPasswordService(passwordString string) (svc *passwordService, err error) {
	var encryptedKey []byte
	salt := make([]byte, saltSize)
	if err = datakey.RandomBytes(&salt); err != nil {
		return nil, err
	}

	var secretKey []byte
	secretKey, err = keyFromPassword([]byte(passwordString), salt)
	if err != nil {
		return nil, err
	}

	// Create a new key
	fileKey, err := datakey.New()
	if err != nil {
		return nil, err
	}
	passwordAes := datakey.NewService(secretKey)

	// encrypt that key with the password hash
	if encryptedKey, err = passwordAes.Encrypt(fileKey); err != nil {
		return
	}

	// Create a service with the encrypted key
	svc = &passwordService{
		encryptedKey: &encryptedKey,
		dataKey:      datakey.NewService(fileKey),
		salt:         &salt,
	}

	return
}

// Hydrate a service from a persisted key
func passwordServiceFromKey(key string) (svc *passwordService, err error) {
	var keyBytes []byte
	if keyBytes, err = base64.StdEncoding.DecodeString(key); err != nil {
		return svc, fmt.Errorf("Could not load password service, crypto.key is not valid base64. %s", err)
	}

	encryptedKey := keyBytes[saltSize:]
	salt := keyBytes[:saltSize]
	svc = &passwordService{
		encryptedKey: &encryptedKey,
		salt:         &salt,
	}
	return
}

// Serialize renders the service key
func (svc *passwordService) Serialize() string {
	// The service key is composed of the salt used to hash the password with
	encodedKey := append([]byte{}, *svc.salt...)
	// followed by the encrypted key
	encodedKey = append(encodedKey, *svc.encryptedKey...)
	return base64.StdEncoding.EncodeToString(encodedKey)
}

// IsAvailable tells if the service has a decrypted key
func (svc *passwordService) IsAvailable() bool {
	return svc.dataKey != nil
}

// DecryptKey decrypts the encrypted key with a given password
func (svc *passwordService) DecryptKey(password string) (err error) {
	if svc.IsAvailable() {
		return
	}

	if svc.encryptedKey == nil {
		return fmt.Errorf("No key found")
	}

	var secretKey []byte
	secretKey, err = keyFromPassword([]byte(password), *svc.salt)
	if err != nil {
		return err
	}

	dataKey, err := datakey.NewService(secretKey).DecryptAsBytes(*svc.encryptedKey)
	if err != nil {
		return fmt.Errorf("Could initialize password service, invalid password: (%d) %s", len(*svc.encryptedKey), err)
	}

	svc.dataKey = datakey.NewService(dataKey)
	return
}

// Decrypt plaintext
func (svc *passwordService) Decrypt(encryptedBytes []byte) (plainText string, err error) {
	return svc.dataKey.Decrypt(encryptedBytes)
}

// Encrypt plaintext
func (svc *passwordService) Encrypt(plainText []byte) (cipherText []byte, err error) {
	return svc.dataKey.Encrypt(plainText)
}
