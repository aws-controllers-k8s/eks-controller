	if desired.ko.Spec.AccessPolicies != nil && len(desired.ko.Spec.AccessPolicies) > 0 {
		// The CreateAccessEntry API does not accept policies, so we must call
		// AssociateAccessPolicy separately after creation.
		latestForSync := &resource{ko.DeepCopy()}
		latestForSync.ko.Spec.AccessPolicies = nil
		if err = rm.syncAccessPolicies(ctx, desired, latestForSync); err != nil {
			return nil, err
		}
	}
