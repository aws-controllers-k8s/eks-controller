	if delta.DifferentAt("Spec.AccessPolicies") {
		err := rm.syncAccessPolicies(ctx, desired, latest)
		if err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Tags") {
		err := rm.syncTags(ctx, desired, latest)
		if err != nil {
			return nil, err
		}
	}
    if !delta.DifferentExcept("Spec.AccessPolicies", "Spec.Tags"){
        return desired, nil
    }