package file

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/blinkhealth/go-config-yourself/internal/yaml"
	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"
	log "github.com/sirupsen/logrus"
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

	hash := sha256.New()
	_, _ = hash.Write(plainText)

	cipherText := base64.StdEncoding.EncodeToString(encryptedBytes)

	return map[string]interface{}{
		"ciphertext": cipherText,
		"encrypted":  true,
		"hash":       fmt.Sprintf("%x", hash.Sum(nil)),
	}, nil
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
