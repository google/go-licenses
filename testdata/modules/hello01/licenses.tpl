{{ range . }}
 - {{.Name}} ([{{.LicenseName}}]({{.LicenseURL}}))
{{- end }}
