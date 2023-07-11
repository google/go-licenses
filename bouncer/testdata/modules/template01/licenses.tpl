{{ range . }}
 - {{.Name}} {{.Version}} ([{{.LicenseName}}]({{.LicenseURL}}))
{{- end }}
