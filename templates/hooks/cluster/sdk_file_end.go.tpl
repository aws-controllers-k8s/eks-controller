{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}

{{/* Find the structure field within the operation */}}
{{- range $fieldName, $field := $CRD.SpecFields -}}
{{- if (or (eq $field.Path "Logging") (eq $field.Path "ResourcesVPCConfig")) }}

{{- $shapeName := $field.ShapeRef.ShapeName }}

// new{{ $shapeName }} returns a {{ $shapeName }} object 
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) new{{ $shapeName }}(
    r *resource,
) *svcsdk.{{ $shapeName }} {
    res := &svcsdk.{{ $shapeName }}{}

{{ GoCodeSetSDKForStruct $CRD "" "res" $field.ShapeRef "" (printf "r.ko.Spec.%s" $field.Names.Camel) 1 }}

    return res
}


{{- end }}

{{- end }}