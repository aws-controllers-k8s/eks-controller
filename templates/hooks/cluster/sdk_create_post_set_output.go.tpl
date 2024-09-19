	if desired.ko.Spec.ResourcesVPCConfig.SubnetRefs != nil {
		ko.Spec.ResourcesVPCConfig.SubnetRefs = desired.ko.Spec.ResourcesVPCConfig.SubnetRefs
	}

	if desired.ko.Spec.ResourcesVPCConfig.SecurityGroupRefs != nil {
		ko.Spec.ResourcesVPCConfig.SecurityGroupRefs = desired.ko.Spec.ResourcesVPCConfig.SecurityGroupRefs
	}
	
	// We expect the cluster to be in 'CREATING' status since we just issued
	// the call to create it, but I suppose it doesn't hurt to check here.
	if clusterCreating(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
		return &resource{ko}, nil
	}

