package Resources

type Effect struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Uid  string `xml:"uid,attr"`
}

func NewEffect() *Effect {
	return &Effect{
		Text: "",
		ID:   "r2",
		Name: "基本字幕",
		Uid:  ".../Titles.localized/Bumper:Opener.localized/Basic Title.localized/Basic Title.moti",
	}
}
