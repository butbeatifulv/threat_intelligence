{{- define "veil.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "veil.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "veil.name" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "veil.observabilityEnv" -}}
- name: OTEL_ENABLED
  value: {{ .Values.observability.enabled | quote }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ .Values.observability.otelEndpoint | quote }}
- name: OTEL_TRACES_SAMPLER_ARG
  value: {{ .Values.observability.samplerRatio | quote }}
- name: LOG_FORMAT
  value: {{ .Values.observability.logFormat | quote }}
- name: LOG_LEVEL
  value: {{ .Values.observability.logLevel | quote }}
{{- end }}
