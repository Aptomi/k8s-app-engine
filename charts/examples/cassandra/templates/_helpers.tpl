{{- define "cassandra.fullname" -}}
{{- printf "cassandra-%s" .Release.Name | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "cassandra.address" -}}
{{ template "cassandra.fullname" . }}.{{ .Release.Namespace }}:{{ .Values.config.ports.cql }}
{{- end -}}
