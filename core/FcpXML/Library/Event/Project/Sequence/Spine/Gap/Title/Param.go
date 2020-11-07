package Title

type Param struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name,attr"`
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

func NewParams(name, key, value string) *Param {
	return &Param{
		Name:  name,
		Key:   key,
		Value: value,
	}
}
