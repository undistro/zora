{{/*
Expand the name of the chart.
*/}}
{{- define "inspect.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "inspect.fullname" -}}
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
{{- define "inspect.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "inspect.labels" -}}
helm.sh/chart: {{ include "inspect.chart" . }}
{{ include "inspect.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Operator labels
*/}}
{{- define "inspect.operatorLabels" -}}
{{ include "inspect.labels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Server labels
*/}}
{{- define "inspect.serverLabels" -}}
{{ include "inspect.labels" . }}
app.kubernetes.io/component: server
{{- end }}

{{/*
UI labels
*/}}
{{- define "inspect.uiLabels" -}}
{{ include "inspect.labels" . }}
app.kubernetes.io/component: ui
{{- end }}

{{/*
NGINX labels
*/}}
{{- define "inspect.nginxLabels" -}}
{{ include "inspect.labels" . }}
app.kubernetes.io/component: nginx
{{- end }}

{{/*
Selector labels
*/}}
{{- define "inspect.selectorLabels" -}}
app.kubernetes.io/name: {{ include "inspect.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Operator selector labels
*/}}
{{- define "inspect.operatorSelectorLabels" -}}
{{ include "inspect.selectorLabels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Server selector labels
*/}}
{{- define "inspect.serverSelectorLabels" -}}
{{ include "inspect.selectorLabels" . }}
app.kubernetes.io/component: server
{{- end }}

{{/*
UI selector labels
*/}}
{{- define "inspect.uiSelectorLabels" -}}
{{ include "inspect.selectorLabels" . }}
app.kubernetes.io/component: ui
{{- end }}

{{/*
NGINX selector labels
*/}}
{{- define "inspect.nginxSelectorLabels" -}}
{{ include "inspect.selectorLabels" . }}
app.kubernetes.io/component: nginx
{{- end }}

{{/*
Create the name of the service account to use in Operator
*/}}
{{- define "inspect.operatorServiceAccountName" -}}
{{- if .Values.operator.rbac.serviceAccount.create }}
{{- default (printf "%s-%s" (include "inspect.fullname" .) "operator") .Values.operator.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.operator.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use in Server
*/}}
{{- define "inspect.serverServiceAccountName" -}}
{{- if .Values.server.rbac.serviceAccount.create }}
{{- default (printf "%s-%s" (include "inspect.fullname" .) "server") .Values.server.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.server.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use in UI
*/}}
{{- define "inspect.uiServiceAccountName" -}}
{{- if .Values.ui.serviceAccount.create }}
{{- default (printf "%s-%s" (include "inspect.fullname" .) "ui") .Values.ui.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.ui.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use in NGINX
*/}}
{{- define "inspect.nginxServiceAccountName" -}}
{{- if .Values.nginx.serviceAccount.create }}
{{- default (printf "%s-%s" (include "inspect.fullname" .) "nginx") .Values.nginx.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.nginx.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "imagePullSecret" }}
{{- with .Values.imageCredentials }}
{{- printf "{\"auths\":{\"%s\":{\"auth\":\"%s\"}}}" .registry (printf "%s:%s" .username .password | b64enc) | b64enc }}
{{- end }}
{{- end }}
