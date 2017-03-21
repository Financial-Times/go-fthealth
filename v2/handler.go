package v2

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"fmt"
)

type checkHandler struct {
	*HealthCheck
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func Handler(hc *HealthCheck) func(w http.ResponseWriter, r *http.Request) {
	ch := &checkHandler{hc}
	return ch.handle
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
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(ErrorMessage{fmt.Sprintf("Failed to encode healthcheck response for % service, erorr was: %v", health.SystemCode, err)})
		w.Write([]byte(msg))
		return
	}
}

func writeHTMLResp(w http.ResponseWriter, health HealthResult) error {
	w.Header().Set("Content-Type", "text/html")
	t := template.New("healthchecks")
	t, err := t.Parse(` <!DOCTYPE html>
                            <head>
                                <title>{{ .Name }} healthchecks </title>
                            </head>

                            <body>
                                <h4>
                                System code: {{ .SystemCode }}<br>
                                Name: {{ .Name }}<br>
                                Description: {{ .Description }}
                                </h4>

                                <h4>Checks:</h4>
                                <ul>
                                    {{ range $key, $value := .Checks }}
                                        <li>
                                            <strong>{{ $value.Name }} </strong>
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
