{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}

{{- define "master-fullname" -}}
{{- printf "spark-master-%s" .Release.Name | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "worker-fullname" -}}
{{- printf "spark-worker-%s" .Release.Name | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "spark-address" -}}
{{- $ctx := . -}}
{{- range $i, $e := until (int .Values.spark.master.replicas) -}}
    {{- if $i }},{{- end -}}
    {{- template "master-fullname" $ctx -}}
    {{- printf "-%d.%s:%d" $e $.Release.Namespace (int $ctx.Values.spark.master.rpcPort) -}}
{{- end -}}
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
        {{- printf .Values.zookeeper.addresses.zookeeper -}}
    {{- end -}}
{{- end -}}
