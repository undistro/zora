{{/*
Expand the name of the chart.
*/}}
{{- define "zora.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "zora.fullname" -}}
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
{{- define "zora.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "zora.labels" -}}
helm.sh/chart: {{ include "zora.chart" . }}
{{ include "zora.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Operator labels
*/}}
{{- define "zora.operatorLabels" -}}
{{ include "zora.labels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Selector labels
*/}}
{{- define "zora.selectorLabels" -}}
app.kubernetes.io/name: {{ include "zora.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Operator selector labels
*/}}
{{- define "zora.operatorSelectorLabels" -}}
{{ include "zora.selectorLabels" . }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Create the name of the service account to use in Operator
*/}}
{{- define "zora.operatorServiceAccountName" -}}
{{- if .Values.operator.rbac.serviceAccount.create }}
{{- default (printf "%s-%s" (include "zora.fullname" .) "operator") .Values.operator.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.operator.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
TokenRefresh selector labels
*/}}
{{- define "zora.tokenRefreshSelectorLabels" -}}
{{ include "zora.selectorLabels" . }}
app.kubernetes.io/component: token-refresh
{{- end }}

{{/*
Create the name of the service account to use in TokenRefresh
*/}}
{{- define "zora.tokenRefreshServiceAccountName" -}}
{{- if .Values.tokenRefresh.rbac.serviceAccount.create }}
{{- default (printf "%s-%s" (include "zora.fullname" .) "token-refresh") .Values.tokenRefresh.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.tokenRefresh.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "zora.imagePullSecret" }}
{{- with .Values.imageCredentials }}
{{- printf "{\"auths\":{\"%s\":{\"auth\":\"%s\"}}}" .registry (printf "%s:%s" .username .password | b64enc) | b64enc }}
{{- end }}
{{- end }}

{{- define "zora.clusterName" }}
{{- include "truncate.name" (dict "name" (regexReplaceAll "\\W+" (required "`clusterName` is required." .Values.clusterName) "-") "len" 63 ) }}
{{- end }}

{{- define "zora.hourlySchedule" }}
{{- $minute := add 5 (now | date "04") }}
{{- if ge $minute 60 }}
{{- $minute = sub $minute 60 }}
{{- end }}
{{- printf "%d * * * *" $minute }}
{{- end }}

{{- define "zora.dailySchedule" }}
{{- $hour := (dateInZone "15" (now) "UTC" | int) }}
{{- $minute := add 5 (now | date "04") }}
{{- if ge $minute 60 }}
{{- $minute = sub $minute 60 }}
{{- $hour = add1 $hour }}
{{- end }}
{{- printf "%d %d * * *" $minute $hour }}
{{- end }}

{{- define "zora.misconfigSchedule" }}
{{- default (include "zora.hourlySchedule" .) .Values.scan.misconfiguration.schedule }}
{{- end }}


{{- define "zora.vulnSchedule" }}
{{- default (include "zora.dailySchedule" .) .Values.scan.vulnerability.schedule }}
{{- end }}

{{/*
Truncate a name to a specific length
@param .name the name of the component
@param .len the maximum length to return
*/}}
{{- define "truncate.name" }}
{{- if gt (len .name) .len }}
{{- $maxLen := int (sub .len 3) }}
{{- $suffixLen := int (div $maxLen 2) }}
{{- $prefixLen := int (sub $maxLen $suffixLen) }}
{{- $suffixStart := int (sub (len .name) $suffixLen) }}
{{- printf "%s---%s" (substr 0 $prefixLen .name) (substr $suffixStart (len .name) .name) }}
{{- else }}
{{- .name }}
{{- end }}
{{- end }}

{{/* Returns true if the explicitly set misconfiguration schedule is more frequently than hourly */}}
{{- define "zora.IsMisconfigScheduleMoreOftenThanHourly" -}}
{{- $cron_fields := split " " .Values.scan.misconfiguration.schedule -}}
{{- $minute := $cron_fields._0 -}}
{{/* minute must be in range [0-59] */}}
{{- not (mustRegexMatch "^(?:\\d|[0-5]\\d)$" $minute) -}}
{{- end -}}

{{/* Returns true if the explicitly set vulnerability schedule is more frequently than daily */}}
{{- define "zora.IsVulnScheduleMoreOftenThanDaily" -}}
{{- $cron_fields := split " " .Values.scan.vulnerability.schedule -}}
{{- $minute := $cron_fields._0 -}}
{{- $hour := $cron_fields._1 -}}
{{/* minute and hour must be in range [0-59] */}}
{{- $isMinuteBad := not (mustRegexMatch "^(?:\\d|[0-5]\\d)$" $minute) -}}
{{- $isHourBad := not (mustRegexMatch "^(?:\\d|[0-5]\\d)$" $hour) -}}
{{- or $isMinuteBad $isHourBad -}}
{{- end -}}

{{- define "zora.saasTokenSecretName" -}}
{{- printf "%s-saas-tokens" (include "zora.fullname" .) -}}
{{- end }}
