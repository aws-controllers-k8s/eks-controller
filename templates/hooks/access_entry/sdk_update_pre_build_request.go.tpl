	// Carry the latest observed status so the update path doesn't return the
	// stale create-time ResourceSynced=False condition (community#2967).
	updatedDesired := rm.concreteResource(desired.DeepCopy())
	updatedDesired.SetStatus(latest)
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
        return updatedDesired, nil
    }
