package normal

type Parser struct {
}

func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (this *Parser) GetParserType() string {
	return "normal"
}
