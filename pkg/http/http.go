package http

import (
	"fmt"
	"html/template"
	"io"
	"strconv"

	pb_almanac "dinowernli.me/almanac/proto"
)

const (
	ingesterHtmlTemplate = "ingester.html.tmpl"
	mixerHtmlTemplate    = "mixer.html.tmpl"
)

var (
	htmlTemplates = template.Must(template.ParseFiles(
		"pkg/http/templates/ingester.html.tmpl",
		"pkg/http/templates/mixer.html.tmpl",
	))
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
	err := htmlTemplates.ExecuteTemplate(writer, ingesterHtmlTemplate, d)
	if err != nil {
		return fmt.Errorf("unable to render template %s: %v", ingesterHtmlTemplate, err)
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
	err := htmlTemplates.ExecuteTemplate(writer, mixerHtmlTemplate, d)
	if err != nil {
		return fmt.Errorf("unable to render template %s: %v", mixerHtmlTemplate, err)
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
