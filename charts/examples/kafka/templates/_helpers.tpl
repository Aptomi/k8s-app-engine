{{- define "kafka-fullname" -}}
{{- printf "kafka-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "zookeeper-fullname" -}}
{{- printf "zk-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "zk-address" -}}
    {{- if .Values.zookeeper.deployChart -}}
        {{- $ctx := . -}}
        {{- range $i, $e := until (int $.Values.zookeeper.replicas) -}}
            {{- if $i }},{{- end -}}
            {{- template "zk-fullname" $ctx -}}
            {{- printf "-%d." $i -}}
            {{- template "zk-fullname" $ctx -}}
            {{- printf ":%d" (int $.Values.zookeeper.clientPort) -}}
         {{- end -}}
    {{- else -}}
        {{- printf "%s" .Values.zookeeper.addresses.zookeeper -}}
    {{- end -}}
{{- end -}}

{{- define "kafka.address" -}}
{{- $ctx := . -}}
{{- range $i, $e := until (int .Values.replicas) -}}
    {{- if $i -}},{{- end -}}
   {{- template "kafka-fullname" $ctx -}}-{{ $i }}.{{- template "kafka-fullname" $ctx -}}.{{ $.Release.Namespace }}:{{ $.Values.port }}
{{- end -}}
{{- end -}}
