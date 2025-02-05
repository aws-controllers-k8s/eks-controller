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
			ToACKTags(desired.ko.Spec.Tags), ToACKTags(latest.ko.Spec.Tags),
		)
		if err != nil {
			return nil, err
		}
	}
    if !delta.DifferentExcept("Spec.AccessPolicies", "Spec.Tags"){
        return desired, nil
    }