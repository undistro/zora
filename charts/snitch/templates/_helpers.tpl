{{/*
Expand the name of the chart.
*/}}
{{- define "snitch.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "snitch.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "snitch.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "snitch.labels" -}}
helm.sh/chart: {{ include "snitch.chart" . }}
{{ include "snitch.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Operator labels
*/}}
{{- define "snitch.operatorLabels" -}}
{{ include "snitch.labels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Server labels
*/}}
{{- define "snitch.serverLabels" -}}
{{ include "snitch.labels" . }}
app.kubernetes.io/component: server
{{- end }}

{{/*
Selector labels
*/}}
{{- define "snitch.selectorLabels" -}}
app.kubernetes.io/name: {{ include "snitch.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Operator selector labels
*/}}
{{- define "snitch.operatorSelectorLabels" -}}
{{ include "snitch.selectorLabels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Server selector labels
*/}}
{{- define "snitch.serverSelectorLabels" -}}
{{ include "snitch.selectorLabels" . }}
app.kubernetes.io/component: server
{{- end }}

{{/*
Create the name of the service account to use in Operator
*/}}
{{- define "snitch.operatorServiceAccountName" -}}
{{- if .Values.operator.serviceAccount.create }}
{{- default (printf "%s-%s" (include "snitch.fullname" .) "operator") .Values.operator.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.operator.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use in Server
*/}}
{{- define "snitch.serverServiceAccountName" -}}
{{- if .Values.server.serviceAccount.create }}
{{- default (printf "%s-%s" (include "snitch.fullname" .) "server") .Values.server.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.server.serviceAccount.name }}
{{- end }}
{{- end }}
