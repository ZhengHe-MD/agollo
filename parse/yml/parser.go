package yml

import (
	"bytes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Parser yml 转换器
type Parser struct {
	Vp *viper.Viper
}

func NewParser() *Parser {
	p := Parser{
		Vp: viper.New(),
	}
	p.Vp.SetConfigType("yml")
	return &p
}

// Parse 内存内容 => yml 数据格式转换器
func (this *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	content, ok := configContent.(string)
	if !ok {
		return nil, nil
	}
	if content == "" {
		return nil, nil
	}

	buffer := bytes.NewBufferString(content)

	err := this.Vp.ReadConfig(buffer)
	if err != nil {
		return nil, err
	}

	return this.convertToMap(), nil
}

func (this *Parser) convertToMap() map[string]interface{} {
	if this.Vp == nil {
		return nil
	}

	m := make(map[string]interface{})
	for _, key := range this.Vp.AllKeys() {
		m[key] = this.Vp.Get(key)
	}
	return m
}

func (this *Parser) GetParserType() string {
	return "yml"
}

func (this *Parser) Unmarshal(data []byte, val interface{}) error {
	return yaml.Unmarshal(data, val)
}
