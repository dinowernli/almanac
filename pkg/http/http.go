package http

import (
	"fmt"
	"html/template"
	"io"
	"strconv"

	"dinowernli.me/almanac/pkg/http/templates"
	pb_almanac "dinowernli.me/almanac/proto"
)

var (
	ingesterTemplate = mustTemplate("ingester.html.tmpl")
	mixerTemplate    = mustTemplate("mixer.html.tmpl")
)

// IngesterData holds the data required to render the ingester page.
type IngesterData struct {
	FormContent string
	Error       error
	Result      string
}

// Render renders the template into the supplied writer using the data
// available in this instance.
func (d *IngesterData) Render(writer io.Writer) error {
	err := ingesterTemplate.Execute(writer, d)
	if err != nil {
		return fmt.Errorf("unable to render ingester template: %v", err)
	}
	return nil
}

// MixerData holds the data required to render the mixer page.
type MixerData struct {
	FormQuery   string
	FormStartMs string
	FormEndMs   string
	Error       error
	Request     *pb_almanac.SearchRequest
	Response    *pb_almanac.SearchResponse
}

// Render renders the template into the supplied writer using the data
// available in this instance.
func (d *MixerData) Render(writer io.Writer) error {
	err := mixerTemplate.Execute(writer, d)
	if err != nil {
		return fmt.Errorf("unable to render mixer template: %v", err)
	}
	return nil
}

// ParseTimestamp parses the input as a timestamp in ms. If parsing fails, this returns
// the supplied fallback.
func ParseTimestamp(input string, fallback int64) int64 {
	if e, err := strconv.Atoi(input); err == nil {
		return int64(e)
	} else {
		return fallback
	}
}

func mustTemplate(assetName string) *template.Template {
	return template.Must(template.New("").Parse(string(templates.MustAsset(assetName))))
}
