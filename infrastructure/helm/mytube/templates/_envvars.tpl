{{/* reusable env var definitions form mytube apps and services */}}
{{- define "mytube.envvars" }}
env:
- name: POSTGRES_HOST
  value: {{ regexReplaceAll ":.*" .Values.db_isntance_endpoint "" }}
- name: API_BASE
  value: http://mytube-service-cluster-ip:3000
{{- end}}
