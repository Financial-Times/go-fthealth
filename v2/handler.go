package v2

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

type checkHandler struct {
	*HealthCheck
}

func (ch *checkHandler) handle(w http.ResponseWriter, r *http.Request) {
	health := ch.health()

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		err := writeHTMLResp(w, health)
		if err == nil {
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err := enc.Encode(health)
	if err != nil {
		panic("write this bit")
	}
}

func Handler(hc *HealthCheck) func(w http.ResponseWriter, r *http.Request) {
	ch := &checkHandler{hc}
	return ch.handle
}

func writeHTMLResp(w http.ResponseWriter, health HealthResult) error {
	w.Header().Set("Content-Type", "text/html")
	t := template.New("healthchecks")
	t, err := t.Parse(` <!DOCTYPE html>
                            <head>
                                <title>{{ .Name }} healthchecks </title>
                            </head>

                            <body>
                                <h2>System code: {{ .SystemCode }}</h2>
                                <h2>Name: {{ .Name }}</h2>
                                <h2>Description: {{ .Description }}</h2>

                                <h2>Checks</h2>
                                <ul>
                                    {{ range $key, $value := .Checks }}
                                        <li>
                                            <strong>Name: {{ $value.Name }} </strong>
                                            <ul>
                                                <li> <strong> Ok: {{ $value.Ok }} </strong> </li>
						<li> Severity: {{ $value.Severity }} </li>
                                                <li> Business impact: {{ $value.BusinessImpact }} </li>
                                                <li> Technical summary: {{ $value.TechnicalSummary }} </li>
						<li> Panic guide: <a href="{{ $value.PanicGuide }}">{{ $value.PanicGuide }}</a> </li>
						<li> Output: {{ $value.CheckOutput }} </li>
                                                <li> Last updated: {{ $value.LastUpdated }} </li>
                                            </ul>
                                        </li>
                                    {{ end }}
                                </ul>
                            </body>`)
	if err == nil {
		t.Execute(w, health)
	}
	return err
}
