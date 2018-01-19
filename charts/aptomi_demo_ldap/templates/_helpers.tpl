{{- define "aptomi_demo_ldap.fullname" -}}
{{- printf "aptomi-demo-ldap-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}