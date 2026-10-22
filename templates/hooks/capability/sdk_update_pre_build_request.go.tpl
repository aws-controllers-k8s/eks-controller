	if err := validateConfiguration(desired); err != nil {
		return nil, err
	}
	// EKS rejects an update while it is still creating, updating or deleting
	// the capability. Requeue instead of spending the call on a
	// ResourceInUseException, and report the state the resource is waiting on.
	if capabilityInFlight(latest) {
		msg := "Capability is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(latest, corev1.ConditionFalse, &msg, nil)
		return latest, requeueWaitUntilCanModify(latest)
	}
	if delta.DifferentAt("Spec.Tags") {
		err := syncTags(
			ctx, rm.sdkapi, rm.metrics, 
			string(*latest.ko.Status.ACKResourceMetadata.ARN), 
			aws.ToStringMap(desired.ko.Spec.Tags), aws.ToStringMap(latest.ko.Spec.Tags),
		)
		if err != nil {
			return nil, err
		}
	}
    if !delta.DifferentExcept("Spec.Tags"){
        return desired, nil
    }
	// A configuration the spec never declared is unmanaged. DescribeCapability
	// reports the configuration block EKS resolved for the capability type, so
	// an undeclared configuration always reads back as a difference, and
	// UpdateCapability has nothing to send for it.
	if desired.ko.Spec.Configuration == nil && !differsOutsideConfiguration(delta) {
		return desired, nil
	}