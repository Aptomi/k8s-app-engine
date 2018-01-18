{{- define "fullname" -}}
{{- printf "aptomi-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}