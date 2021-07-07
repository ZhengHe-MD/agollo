package normal

type Parser struct {
}

func NewParser() *Parser {
	p := Parser{}
	return &p
}

func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (this *Parser) GetParserType() string {
	return "normal"
}

func (this *Parser) Unmarshal(data []byte, val interface{}) error {
	return nil
}
