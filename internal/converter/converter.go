package converter

import (
	"gopkg.in/yaml.v3"
	"strings"
)

func FlattenYAML(yamlData []byte) (map[string]string, map[string]string, map[string]string, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlData, &data)
	if err != nil {
		return nil, nil, nil, err
	}

	flatMap := make(map[string]string)
	uppercaseMap := make(map[string]string)
	templateMap := make(map[string]string)

	flatten("", data, flatMap, uppercaseMap, templateMap)

	return flatMap, uppercaseMap, templateMap, nil
}

func flatten(prefix string, nestedMap map[string]interface{}, flatMap, uppercaseMap, templateMap map[string]string) {
	for key, value := range nestedMap {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[interface{}]interface{}:
			subMap := make(map[string]interface{})
			for k, val := range v {
				subMap[k.(string)] = val
			}
			flatten(fullKey, subMap, flatMap, uppercaseMap, templateMap)
		case map[string]interface{}:
			flatten(fullKey, v, flatMap, uppercaseMap, templateMap)
		default:
			flatMap[fullKey] = value.(string)
			upperKey := strings.ToUpper(strings.ReplaceAll(fullKey, ".", "_"))
			uppercaseMap[upperKey] = value.(string)
			templateMap[fullKey] = "${" + upperKey + "}"
		}
	}
}
