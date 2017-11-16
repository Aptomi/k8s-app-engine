{{- define "zk-fullname" -}}
{{- printf "zk-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "zookeeper.address" -}}
{{- $ctx := . -}}
{{- range $i, $e := until (int $.Values.replicas) -}}
  {{- if $i }},{{- end -}}
  {{- template "zk-fullname" $ctx -}}
  {{- printf "-%d." $i -}}
  {{- template "zk-fullname" $ctx -}}
  {{- printf ".%s:%d" $.Release.Namespace (int $.Values.clientPort) -}}
{{- end -}}
{{- end -}}
