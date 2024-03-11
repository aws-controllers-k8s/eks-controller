{{/* The name of the application this chart installs */}}
{{- define "ack-eks-controller.app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ack-eks-controller.app.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* The name and version as used by the chart label */}}
{{- define "ack-eks-controller.chart.name-version" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* The name of the service account to use */}}
{{- define "ack-eks-controller.service-account.name" -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}

{{- define "ack-eks-controller.watch-namespace" -}}
{{- if eq .Values.installScope "namespace" -}}
{{ .Values.watchNamespace | default .Release.Namespace }}
{{- end -}}
{{- end -}}

{{/* The mount path for the shared credentials file */}}
{{- define "ack-eks-controller.aws.credentials.secret_mount_path" -}}
{{- "/var/run/secrets/aws" -}}
{{- end -}}

{{/* The path the shared credentials file is mounted */}}
{{- define "ack-eks-controller.aws.credentials.path" -}}
{{ $secret_mount_path := include "ack-eks-controller.aws.credentials.secret_mount_path" . }}
{{- printf "%s/%s" $secret_mount_path .Values.aws.credentials.secretKey -}}
{{- end -}}

{{/* The rules a of ClusterRole or Role */}}
{{- define "ack-eks-controller.rbac-rules" -}}
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - patch
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - patch
  - watch
- apiGroups:
  - ec2.services.k8s.aws
  resources:
  - securitygroups
  verbs:
  - get
  - list
- apiGroups:
  - ec2.services.k8s.aws
  resources:
  - securitygroups/status
  verbs:
  - get
  - list
- apiGroups:
  - ec2.services.k8s.aws
  resources:
  - subnets
  verbs:
  - get
  - list
- apiGroups:
  - ec2.services.k8s.aws
  resources:
  - subnets/status
  verbs:
  - get
  - list
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - accessentries
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - accessentries/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - addons
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - addons/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - clusters
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - clusters/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - fargateprofiles
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - fargateprofiles/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - nodegroups
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - nodegroups/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - podidentityassociations
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - eks.services.k8s.aws
  resources:
  - podidentityassociations/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - iam.services.k8s.aws
  resources:
  - roles
  verbs:
  - get
  - list
- apiGroups:
  - iam.services.k8s.aws
  resources:
  - roles/status
  verbs:
  - get
  - list
- apiGroups:
  - kms.services.k8s.aws
  resources:
  - keys
  verbs:
  - get
  - list
- apiGroups:
  - kms.services.k8s.aws
  resources:
  - keys/status
  verbs:
  - get
  - list
- apiGroups:
  - services.k8s.aws
  resources:
  - adoptedresources
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - services.k8s.aws
  resources:
  - adoptedresources/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - services.k8s.aws
  resources:
  - fieldexports
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - services.k8s.aws
  resources:
  - fieldexports/status
  verbs:
  - get
  - patch
  - update
{{- end }}