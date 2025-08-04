// UpdatePodIdentityAssociation does not unset TargetRoleARN if input set to nil.
// Need to provide empty string instead for TargetRoleARN to be unset in update operation.
if desired.ko.Spec.TargetRoleARN == nil {
    temp := ""
    input.TargetRoleArn = &temp
}