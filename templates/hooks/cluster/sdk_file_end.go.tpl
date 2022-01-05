{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}

{{- $updateClusterVersion := (index $SDKAPI.API.Operations "UpdateClusterConfig") -}}

{{/* Find the structure field within the operation */}}
{{- range $memberRefName, $memberRef := $updateClusterVersion.InputRef.Shape.MemberRefs -}}
{{- if (or (eq $memberRefName "Logging") (eq $memberRefName "ResourcesVpcConfig")) }}

// new{{ $memberRefName }} returns a {{ $memberRefName }} object 
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) new{{ $memberRefName }}(
    r *resource,
) *svcsdk.{{ $memberRef.ShapeName }} {
    res := &svcsdk.{{ $memberRef.ShapeName }}{}

{{ $names := (ToNames $memberRefName) -}}
{{ GoCodeSetSDKForStruct $CRD "" "res" $memberRef "" (printf "r.ko.Spec.%s" $names.Camel) 1 }}

    return res
}


{{- end }}

{{- end }}