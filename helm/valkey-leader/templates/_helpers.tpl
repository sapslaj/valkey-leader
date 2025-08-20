{{/*
Expand the name of the chart.
*/}}
{{- define "valkey-leader.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "valkey-leader.fullname" -}}
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
{{- define "valkey-leader.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "valkey-leader.labels" -}}
helm.sh/chart: {{ include "valkey-leader.chart" . }}
{{ include "valkey-leader.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "valkey-leader.selectorLabels" -}}
app.kubernetes.io/name: {{ include "valkey-leader.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Valkey cluster labels
*/}}
{{- define "valkey-leader.clusterLabels" -}}
valkey.sapslaj.cloud/cluster: {{ include "valkey-leader.clusterName" . }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "valkey-leader.serviceAccountName" -}}
{{- if .Values.rbac.serviceAccount.create }}
{{- default (include "valkey-leader.fullname" .) .Values.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the cluster name
*/}}
{{- define "valkey-leader.clusterName" -}}
{{- default .Release.Name .Values.cluster.name }}
{{- end }}

{{/*
Create the leader lease name
*/}}
{{- define "valkey-leader.leaseName" -}}
{{- default .Release.Name .Values.leaderElection.leaseName }}
{{- end }}
