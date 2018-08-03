package v1_1

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type checkHandler struct {
	HC
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func Handler(hc HC) func(w http.ResponseWriter, r *http.Request) {
	ch := checkHandler{hc}
	return ch.handle
}

func (ch *checkHandler) handle(w http.ResponseWriter, r *http.Request) {
	health := RunCheck(ch)

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
		<style>
			h3 {
				padding: 0.5em 1em;
				display: inline-block;
				border-radius: 0.5em;
				margin: 0;
			}
			.ok {
				background-color: #458b00;
				color: #fff;
			}
			.error {
				background-color: #b00;
				color: #fff;
			}
			.output {
				background: #ccc;
				border: solid thin #999;
				padding: 0.5em;
			}
		</style>
	</head>

	<body>
		<h1>Healthcheck for {{ .Name }}</h1>
		<table>
			<tr><th>Description</th><td>{{ .Description }}</td></tr>
			<tr><th>System Code</th><td>{{ .SystemCode }}</td></tr>
			{{if .SystemCode }}<tr>
				<th>Runbook</th>
				<td><a href="https://dewey.in.ft.com/runbooks/{{ .SystemCode }}" target="__blank">https://dewey.in.ft.com/runbooks/{{ .SystemCode }}</a></td>
			</tr>{{ end }}
		</table>

		<h2>Checks</h2>
			{{ range $key, $value := .Checks }}
				<h3 class="{{if $value.Ok }}ok{{ else }}error{{ end }}">{{ $value.Name }}</h3>
				<ul>
					<li> Status: {{if $value.Ok }}OK{{ else }}Error{{ end }}</li>
					<li> Severity: {{ $value.Severity }} </li>
					<li> Business impact: {{ $value.BusinessImpact }} </li>
					<li> Technical summary: {{ $value.TechnicalSummary }} </li>
					{{if $value.PanicGuideIsLink }}<li> Panic guide: <a href="{{ $value.PanicGuide }}">{{ $value.PanicGuide }}</a> </li>
					{{ else }}<li> Panic guide: <pre>{{ $value.PanicGuide }}</pre> </li>{{ end }}
					{{if $value.CheckOutput }}<li> Output: <pre class='output'>{{ $value.CheckOutput }}</pre> </li>{{ end }}
					<li> Last updated: {{ $value.LastUpdated }} </li>
				</ul>
			{{ end }}
	</body>`)
	if err == nil {
		t.Execute(w, health)
	}
	return err
}
