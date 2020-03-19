{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kms-secrets.name" -}}
{{- default .Chart.Name .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "kms-secrets.serviceAccountName" -}}
{{- $name := default "default" .Values.rbac.serviceAccount.name }}
{{- printf "%s-%s" (include "kms-secrets.name" .) $name }}
{{- end -}}
