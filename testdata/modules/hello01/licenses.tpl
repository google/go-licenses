{{ range . }}
## {{.Name}} ([{{.LicenseName}}]({{.LicenseURL}}))

```
{{- licenseText . -}}
```
{{- end }}
