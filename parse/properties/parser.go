package properties

import (
	"github.com/spf13/viper"
)

// Parser properties 转换器
type Parser struct {
	Vp *viper.Viper
}

func NewParser() *Parser {
	p := Parser{
		Vp: viper.New(),
	}
	p.Vp.SetConfigType("properties")
	return &p
}

// Parse 内存内容 => properties 数据格式转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (this *Parser) GetParserType() string {
	return "properties"
}

func (this *Parser) Unmarshal(data []byte, val interface{}) error {
	return nil
}
