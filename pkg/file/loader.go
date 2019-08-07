package file

import (
	"fmt"
	"strings"

	"github.com/blinkhealth/go-config-yourself/internal/yaml"
	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"

	// register gpg
	_ "github.com/blinkhealth/go-config-yourself/pkg/crypto/gpg"
	// register kms
	_ "github.com/blinkhealth/go-config-yourself/pkg/crypto/kms"
	// register password
	_ "github.com/blinkhealth/go-config-yourself/pkg/crypto/password"

	log "github.com/sirupsen/logrus"
)

// Create a new ConfigFile, initializing its crypto provider with given arguments.
//
// The user may be prompted for details if connected to a TTY and these are not provided by `providerArgs`
func Create(providerName string, providerArgs map[string]interface{}) (config *ConfigFile, err error) {
	cfgMap := make(map[string]interface{})
	pvd, err := initializeProvider(providerName, providerArgs)

	if err != nil {
		return nil, err
	}

	err = pvd.Replace(providerArgs)
	if err != nil {
		return
	}

	cfgMap["crypto"] = pvd.Serialize()

	data, _ := yaml.FromValue(cfgMap)

	config = &ConfigFile{
		data:     data,
		crypto:   pvd,
		Provider: providerName,
	}
	return
}

// Load a file at a give path and return a ConfigFile
func Load(path string) (config *ConfigFile, err error) {
	data, err := yaml.FromPathname(path)
	if err != nil {
		return nil, fmt.Errorf("Could not parse YAML: %s", err)
	}

	var provider pvd.Crypto
	var providerName string

	cryptoMap := map[string]interface{}{}
	if err = data.Get("crypto", &cryptoMap); err == nil {
		iProvider, providerDefined := cryptoMap["provider"]
		if !providerDefined {
			log.Warn("Setting default provider to kms")
			providerName = "kms"
		} else {
			providerName = iProvider.(string)
		}

		log.Debugf("Initializing secure config with provider: %s", providerName)

		provider, err = initializeProvider(providerName, cryptoMap)
		if err != nil {
			return
		}
	} else {
		errString := err.Error()
		if strings.Contains(errString, "yaml: unmarshal errors") {
			err = fmt.Errorf("Invalid config, the crypto property is not a map: %s", err)
			return
		}
		err = nil
	}

	config = &ConfigFile{
		data:     data,
		crypto:   provider,
		Provider: providerName,
	}

	return
}

func initializeProvider(providerName string, config map[string]interface{}) (pvd.Crypto, error) {
	provider, ok := pvd.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("Unknown provider <%s>", providerName)
	}

	return provider.New(config)
}
