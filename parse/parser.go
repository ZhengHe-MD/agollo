package parse

import (
	"github.com/ZhengHe-MD/agollo/v4/parse/json"
	"github.com/ZhengHe-MD/agollo/v4/parse/normal"
	"github.com/ZhengHe-MD/agollo/v4/parse/properties"
	"github.com/ZhengHe-MD/agollo/v4/parse/yaml"
	"github.com/ZhengHe-MD/agollo/v4/parse/yml"
)

type ContentParser interface {
	Parse(configContent interface{}) (map[string]interface{}, error)
	GetParserType() string
	Unmarshal(data []byte, val interface{}) error
}

func GetParser(typ string) ContentParser {
	switch typ {
	case "yml":
		return yml.NewParser()
	case "yaml":
		return yaml.NewParser()
	case "json":
		return json.NewParser()
	case "properties":
		return properties.NewParser()
	default:
		return normal.NewParser()
	}
}
