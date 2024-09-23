package devops_go_vault_api

import (
	"bufio"
	"os"
	"strings"
)

func ParseIniFile(filename string) (map[string]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]map[string]string)
	var section string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			config[section] = make(map[string]string)
		} else {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[section][key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func TransformToJson1(config map[string]map[string]string) map[string]string {
	result := make(map[string]string)
	for section, items := range config {
		for key, _ := range items {
			jsonKey := key
			if section != "" {
				jsonKey = strings.ToUpper(section) + "_" + strings.ReplaceAll(strings.ToUpper(key), ".", "_")
			}
			result[key] = "${" + jsonKey + "}"
		}
	}
	return result
}

func TransformToJson2(config map[string]map[string]string) map[string]string {
	result := make(map[string]string)
	for section, items := range config {
		for key, value := range items {
			jsonKey := strings.ToUpper(key)
			if section != "" {
				jsonKey = strings.ToUpper(section) + "_" + strings.ReplaceAll(jsonKey, ".", "_")
			}
			result[jsonKey] = value
		}
	}
	return result
}
