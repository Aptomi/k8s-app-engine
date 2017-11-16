{{- define "namenode-fullname" -}}
{{- printf "hdfs-namenode-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "datanode-fullname" -}}
{{- printf "hdfs-datanode-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "configmap-fullname" -}}
{{- printf "hdfs-configs-%s" .Release.Name | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "hdfs-ui-fullname" -}}
{{- printf "hdfs-ui-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "namenode-address" -}}
{{- template "namenode-fullname" . }}-0.{{- template "namenode-fullname" . }}.{{ .Release.Namespace }}:{{ .Values.namenode.port }}
{{- end -}}

{{- define "hdfs-ui-address" -}}
{{ template "hdfs-ui-fullname" . }}.{{ .Release.Namespace }}:{{ .Values.namenode.ui.port }}
{{- end -}}
