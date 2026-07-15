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

{{- define "veil.playbookEnv" -}}
- name: VEIL_REPO_ROOT
  value: {{ .Values.playbooks.mountPath | quote }}
- name: VEIL_CYBER_SKILLS_INDEX
  value: {{ printf "%s/docs/skills-index/cyber-skills.json" .Values.playbooks.mountPath | quote }}
- name: VEIL_BLEVE_INDEX_PATH
  value: {{ printf "%s/docs/skills-index/playbook-search.bleve" .Values.playbooks.mountPath | quote }}
- name: VEIL_SEARCH_ENGINE
  value: {{ .Values.playbooks.searchEngine | quote }}
- name: VEIL_SEARCH_MODE
  value: {{ .Values.playbooks.searchMode | quote }}
{{- end }}

{{- define "veil.playbookVolume" -}}
- name: veil-playbooks
  hostPath:
    path: {{ .Values.playbooks.hostPath | quote }}
    type: Directory
{{- end }}

{{- define "veil.playbookVolumeMount" -}}
- name: veil-playbooks
  mountPath: {{ .Values.playbooks.mountPath | quote }}
  readOnly: true
{{- end }}
