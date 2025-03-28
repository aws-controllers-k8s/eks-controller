	if delta.DifferentAt("Spec.AccessPolicies") {
		err := rm.syncAccessPolicies(ctx, desired, latest)
		if err != nil {
			return nil, err
		}
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
    if !delta.DifferentExcept("Spec.AccessPolicies", "Spec.Tags"){
        return desired, nil
    }