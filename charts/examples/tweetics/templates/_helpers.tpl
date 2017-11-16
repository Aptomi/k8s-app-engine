{{- define "tweetics.fullname" -}}
{{- printf "tweetics-%s" .Release.Name  | trunc 55 | trimSuffix "-" -}}
{{- end -}}

{{- define "tweetics.kafka-address" -}}
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

{{- define "tweetics.zk-address" -}}
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

{{- define "tweetics.spark-address" -}}
    {{- if .Values.spark.deployChart -}}
        {{- $ctx := . -}}
        {{- range $i, $e := until (int .Values.spark.spark.master.replicas) -}}
            {{- if $i }},{{- end -}}
            {{- template "master-fullname" $ctx -}}
            {{- printf "-%d:%d" $i (int $ctx.Values.spark.spark.master.rpcPort) -}}
        {{- end -}}
    {{- else -}}
        {{- printf "%s" .Values.spark.addresses.spark -}}
    {{- end -}}
{{- end -}}

{{- define "tweetics.hdfs-address" -}}
    {{- if .Values.hdfs.deployChart -}}
        {{- template "namenode-fullname" . -}}-0.{{- template "namenode-fullname" . -}}:{{ .Values.hdfs.namenode.port }}
    {{- else -}}
        {{- printf "%s" .Values.hdfs.addresses.namenode -}}
    {{- end -}}
{{- end -}}

{{- define "tweetics.cassandra-address" -}}
    {{- if .Values.cassandra.deployChart -}}
        {{- printf "cassandra-%s" .Release.Name | trunc 55 | trimSuffix "-" -}}
    {{- else -}}
        {{- (split ":" .Values.cassandra.addresses.cassandra)._0 -}}
    {{- end -}}
{{- end -}}

{{- define "tweetics-storage" -}}
    {{- if eq .Values.storage "hdfs" -}}
        {{- printf "hdfs://" -}}{{ template "tweetics.hdfs-address" . }}{{- .Values.hdfs.path -}}
    {{- else -}}
        {{ template "tweetics.cassandra-address" . }}:{{- .Values.cassandra.keyspace -}}:{{- .Values.cassandra.table -}}
    {{- end -}}
{{- end -}}
