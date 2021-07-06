package parse

import (
	"github.com/ZhengHe-MD/agollo/v4/parse/yaml"
	"github.com/ZhengHe-MD/agollo/v4/parse/yml"
)

type ContentParser interface {
	Parse(configContent interface{}) (map[string]interface{}, error)
	GetParserType() string
}

func GetParser(typ string) ContentParser {
	switch typ {
	case "yml":
		return yml.NewParser()
	case "yaml":
		return yaml.NewParser()
	default:
		return nil
	}
}
