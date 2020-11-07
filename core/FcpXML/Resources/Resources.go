package Resources

type Resources struct {
	Text   string        `xml:",chardata"`
	Format *RenderFormat `xml:"format"`
	Effect *Effect       `xml:"effect"`
}

func NewResources() *Resources {
	return &Resources{
		Text:   "",
		Format: &RenderFormat{},
		Effect: &Effect{},
	}
}

func (r *Resources) SetFormat(format *RenderFormat) *Resources {
	r.Format = format
	return r
}

func (r *Resources) SetEffect(effect *Effect) *Resources {
	r.Effect = effect
	return r
}
