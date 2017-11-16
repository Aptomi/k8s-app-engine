{{- define "fullname" -}}
{{- printf "tweepub-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "kafka-address" -}}
    {{- if .Values.kafka.deployChart -}}
        {{- $ctx := . -}}
        {{- range $i, $e := until (int $.Values.kafka.replicas) -}}
            {{- if $i }},{{- end -}}
            {{- template "kafka-fullname" $ctx -}}
            {{- printf "-%d." $i -}}
            {{- template "kafka-fullname" $ctx -}}
            {{- printf ":%d" (int $.Values.kafka.port) -}}
        {{- end -}}
    {{- else -}}
        {{- printf "%s" .Values.kafka.addresses.kafka -}}
    {{- end -}}
{{- end -}}

{{- define "tweepub.hdfs-address" -}}
    {{- if .Values.hdfs.deployChart -}}
        {{- template "namenode-fullname" . -}}-0.{{- template "namenode-fullname" . -}}:{{ .Values.hdfs.namenode.ui.port }}
    {{- else -}}
        {{- printf "%s" .Values.hdfs.addresses.namenode -}}
    {{- end -}}
{{- end -}}
