package main

import (
	"encoding/xml"
	"fmt"
	"io"
)

type Root struct {
	XmlName     xml.Name `xml:"items"`
	BurpVersion string   `xml:"burpVersion,attr"`
	ExportTime  string   `xml:"exportTime,attr"`

	Items []Item `xml:"item"`
}

type Item struct {
	Time           string      `xml:"time"`
	Url            string      `xml:"url"`
	Host           Host        `xml:"host"`
	Port           int         `xml:"port"`
	Protocol       string      `xml:"protocol"`
	Method         string      `xml:"method"`
	Path           string      `xml:"path"`
	Extension      string      `xml:"extension"`
	Request        BodyPayload `xml:"request"`
	Status         int         `xml:"status"`
	ResponseLength int         `xml:"responselength"`
	MimeType       string      `xml:"mimetype"`
	Response       BodyPayload `xml:"response"`
	Comment        string      `xml:"comment"`
}

type Host struct {
	Ip    string `xml:"ip,attr"`
	Value string `xml:",chardata"`
}

type BodyPayload struct {
	IsBase64 bool   `xml:"base64,attr"`
	Data     string `xml:",cdata"`
}

func (rt *Root) Deserialize(r io.Reader) error {
	decoder := xml.NewDecoder(r)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error getting token: %w", err)
		}

		elem, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch elem.Name.Local {
		case "items":
			for _, attr := range elem.Attr {
				switch attr.Name.Local {
				case "burpVersion":
					rt.BurpVersion = attr.Value
				case "exportTime":
					rt.ExportTime = attr.Value
				}
			}
		case "item":
			var item Item
			if err := decoder.DecodeElement(&item, &elem); err != nil {
				return fmt.Errorf("error decoding item element: %w", err)
			}

			rt.Items = append(rt.Items, item)
		}
	}

	if rt.Items == nil || len(rt.Items) == 0 {
		return fmt.Errorf("no `items` tag, not a Burp sitemap export file?")
	}

	return nil
}
