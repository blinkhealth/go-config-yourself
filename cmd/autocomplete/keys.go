package autocomplete

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/blinkhealth/go-config-yourself/internal/yaml"
	log "github.com/sirupsen/logrus"
)

// AutocompleteKeys enumerates possible subPaths for a given keyPath
func possibleSubKeys(keyPath string, cfg *yaml.Tree) (list []string, err error) {
	var value *yaml.Tree
	var query string

	if keyPath != "" {
		parentPath := strings.Split(keyPath, ".")
		if len(parentPath) > 0 {
			// offer suggestions based on this query
			query = parentPath[len(parentPath)-1]
			if _, err := strconv.Atoi(query); err != nil {
				parentPath = parentPath[:len(parentPath)-1]
			}
			log.Debugf("Querying for %s in %s", query, parentPath)
		}

		if len(parentPath) > 0 {
			err = cfg.Get(strings.Join(parentPath, "."), &value)
			if err != nil {
				return nil, err
			}
		} else {

			value = cfg
		}
	} else {
		value = cfg
	}

	log.Debugf("Query: %s", query)
	if value.IsSlice() {
		theList := make([]interface{}, 0)
		if err = value.Decode(&theList); err != nil {
			return
		}

		for keyInt, value := range theList {
			key := mappedKey(strconv.Itoa(keyInt), value)
			list = append(list, key)
		}
	} else if value.IsMap() {
		theMap := map[string]interface{}{}
		if err = value.Decode(&theMap); err != nil {
			return
		}
		for key, value := range theMap {
			if query != "" && !strings.HasPrefix(strings.ToLower(key), strings.ToLower(query)) {
				log.Debugf("Skipping %s", key)
				continue
			}

			key = mappedKey(key, value)
			list = append(list, key)
		}
	}

	sort.Strings(list)
	return
}

func mappedKey(key string, value interface{}) string {
	switch theValue := value.(type) {
	case map[interface{}]interface{}:
		stringMap := map[string]interface{}{}
		for k, v := range theValue {
			stringKey := k.(string)
			stringMap[stringKey] = v
		}
		if _, hasCrypto := stringMap["encrypted"]; !hasCrypto {
			key = fmt.Sprintf("%s.", key)
		}
	case map[string]interface{}:
		if _, hasCrypto := theValue["encrypted"]; !hasCrypto {
			key = fmt.Sprintf("%s.", key)
		}
	case []interface{}:
		key = fmt.Sprintf("%s.", key)
	default:
		log.Debugf("value is %T", value)
	}

	return key
}
