package file

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/blinkhealth/go-config-yourself/internal/yaml"
	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
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

type cryptoDisabledError struct{}

func (cryptoDisabledError) Error() string {
	return "Unable to decrypt, config file has no `crypto` property, or the crypto provider is not enabled"
}

func decryptNode(node *yaml.Tree, provider pvd.Crypto) (interface{}, error) {
	if node.IsMap() {
		if node.IsEncrypted() {
			if provider == nil || !provider.Enabled() {
				return nil, cryptoDisabledError{}
			}

			plainText, err := provider.Decrypt(*node.Secret)

			if err != nil {
				return nil, err
			}

			return plainText, nil
		}

		outerMap := map[string]*yaml.Tree{}
		retMap := map[string]interface{}{}
		if err := node.Decode(&outerMap); err != nil {
			return nil, err
		}

		for key, value := range outerMap {
			decryptedValue, err := decryptNode(value, provider)
			if err != nil {
				return nil, err
			}
			retMap[key] = decryptedValue
		}

		return retMap, nil
	}

	var value interface{}
	err := node.Decode(&value)
	return value, err
}

func encryptCipherText(plainText []byte, provider pvd.Crypto) (map[string]interface{}, error) {
	log.Debugf("encrypting %d bytes", len(plainText))
	encryptedBytes, err := provider.Encrypt(plainText)

	if err != nil {
		return nil, err
	}

	cipherText := base64.StdEncoding.EncodeToString(encryptedBytes)
	hash, hashingErr := plainTextHash(plainText, provider)
	if hashingErr != nil {
		return nil, hashingErr
	}

	return map[string]interface{}{
		"ciphertext": cipherText,
		"encrypted":  true,
		"hash":       fmt.Sprintf("%x", hash),
	}, nil
}

func plainTextHash(plainText []byte, provider pvd.Crypto) (hash []byte, err error) {
	salt := new(bytes.Buffer)
	enc := gob.NewEncoder(salt)
	err = enc.Encode(provider.Serialize())
	if err != nil {
		return nil, err
	}

	hasher := sha512.New512_256()
	_, _ = hasher.Write(plainText)
	hash, err = scrypt.Key(hasher.Sum(nil), salt.Bytes(), scryptCost, scryptR, scryptP, 32)
	if err != nil {
		return nil, err
	}
	return
}

func secretsForNode(node *yaml.Tree, parent string) []string {
	secrets := make([]string, 0)

	if node.IsMap() {
		if node.IsEncrypted() {
			secrets = append(secrets, parent)
		} else {
			theMap := map[string]*yaml.Tree{}
			if err := node.Decode(&theMap); err != nil {
				panic(err)
			}

			for k, v := range theMap {
				path := k
				if parent != "" {
					path = fmt.Sprintf("%s.%s", parent, k)
				}
				secrets = append(secrets, secretsForNode(v, path)...)
			}
		}
	}

	return secrets
}
