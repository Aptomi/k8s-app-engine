{{- define "fullname" -}}
{{- printf "aptomi-demo-ldap-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}