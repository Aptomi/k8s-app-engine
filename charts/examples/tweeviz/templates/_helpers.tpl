{{- define "fullname" -}}
{{- printf "tweeviz-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}
