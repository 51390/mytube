{{/* reusable env var definitions form mytube apps and services */}}
{{- define "mytube.envvars" }}
env:
- name: POSTGRES_HOST
  value: {{ regexReplaceAll ":.*" .Values.db_isntance_endpoint "" }}
- name: API_BASE
  value: http://mytube-service-cluster-ip:3000
- name: GOOGLE_REDIRECT_URI
  value: {{ .Values.google_redirect_uri }}
- name: REDIRECT_URIS
  value: {{ .Values.redirect_uris }}
- name: APP_ENDPOINT
  value: {{ .Values.app_endpoint }}
- name: APP_PORT
  value: "{{ .Values.app_port }}"
- name: POSTGRES_SSLMODE
  value: {{ .Values.postgres_sslmode }}
{{- end}}
