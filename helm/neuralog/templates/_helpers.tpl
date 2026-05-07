{{/*
Fully qualified name — max 63 chars per DNS spec.
*/}}
{{- define "neuralog.fullname" -}}
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
Chart label: name-version.
*/}}
{{- define "neuralog.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Standard labels applied to every resource.
*/}}
{{- define "neuralog.labels" -}}
helm.sh/chart: {{ include "neuralog.chart" . }}
{{ include "neuralog.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels — must be immutable across upgrades.
*/}}
{{- define "neuralog.selectorLabels" -}}
app.kubernetes.io/name: {{ default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name.
*/}}
{{- define "neuralog.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "neuralog.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Target namespace — allows overriding the Release namespace.
*/}}
{{- define "neuralog.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride }}
{{- end }}

{{/*
Collector image tag (falls back to appVersion).
*/}}
{{- define "neuralog.collectorImage" -}}
{{- printf "%s:%s" .Values.image.collector.repository (default .Chart.AppVersion .Values.image.collector.tag) }}
{{- end }}

{{/*
UI image tag (falls back to appVersion).
*/}}
{{- define "neuralog.uiImage" -}}
{{- printf "%s:%s" .Values.image.ui.repository (default .Chart.AppVersion .Values.image.ui.tag) }}
{{- end }}
