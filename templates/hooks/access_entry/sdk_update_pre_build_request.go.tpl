	// Carry the latest observed status (including conditions) onto a copy of
	// desired so the returned resource reflects the observed state rather than
	// the stale create-time "Access policy update pending" ResourceSynced=False
	// condition. AccessPolicies are applied via AssociateAccessPolicy side-effect
	// calls (not UpdateAccessEntry); returning bare `desired` here leaves that
	// stale condition in place and the resource stays Synced=False until the next
	// full resync (up to 10h). See aws-controllers-k8s/community#2967.
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
