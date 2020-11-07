package FcpXML

import (
	"encoding/xml"
	"srt2fcpxml/core/FcpXML/Library"
	"srt2fcpxml/core/FcpXML/Resources"
)

type FcpXML struct {
	XMLName   xml.Name             `xml:"fcpxml"`
	Text      string               `xml:",chardata"`
	Version   string               `xml:"version,attr"`
	Resources *Resources.Resources `xml:"resources"`
	Library   *Library.Library     `xml:"library"`
}

func New() *FcpXML {
	return &FcpXML{
		XMLName:   xml.Name{},
		Text:      "",
		Version:   "1.7",
		Resources: &Resources.Resources{},
		Library:   &Library.Library{},
	}
}

func (fcp *FcpXML) SetVersion(version string) *FcpXML {
	fcp.Version = version
	return fcp
}

func (fcp *FcpXML) SetResources(resources *Resources.Resources) *FcpXML {
	fcp.Resources = resources
	return fcp
}

func (fcp *FcpXML) SetLibrary(library *Library.Library) *FcpXML {
	fcp.Library = library
	return fcp

}
