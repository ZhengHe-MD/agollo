package properties

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容=>properties文件转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (this *Parser) GetParserType() string {
	return "properties"
}
