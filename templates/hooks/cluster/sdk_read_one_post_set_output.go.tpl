	if r.ko.Spec.ResourcesVPCConfig.SubnetRefs != nil {
		ko.Spec.ResourcesVPCConfig.SubnetRefs = r.ko.Spec.ResourcesVPCConfig.SubnetRefs
	}

	if r.ko.Spec.ResourcesVPCConfig.SecurityGroupRefs != nil {
		ko.Spec.ResourcesVPCConfig.SecurityGroupRefs = r.ko.Spec.ResourcesVPCConfig.SecurityGroupRefs
	}
	
	if !clusterActive(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	} else {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
	}