	if r.ko.Spec.ResourcesVPCConfig != nil && r.ko.Spec.ResourcesVPCConfig.SubnetRefs != nil {
		ko.Spec.ResourcesVPCConfig.SubnetRefs = r.ko.Spec.ResourcesVPCConfig.SubnetRefs
	}

	if r.ko.Spec.ResourcesVPCConfig != nil && r.ko.Spec.ResourcesVPCConfig.SecurityGroupRefs != nil {
		ko.Spec.ResourcesVPCConfig.SecurityGroupRefs = r.ko.Spec.ResourcesVPCConfig.SecurityGroupRefs
	}

	desiredConfig := r.ko.Spec.KubernetesNetworkConfig
	latestConfig := ko.Spec.KubernetesNetworkConfig

	// ElasticLoadBalancing can by default be initialized as false even when ACK is providing an empty input.
	// This condition prevents unnecessary deltas when the desired value is empty and ElasticLoadBalancing is already disabled.
	if desiredConfig != nil && desiredConfig.ElasticLoadBalancing == nil && latestConfig != nil {
		latestConfig.ElasticLoadBalancing = nil
	}
	
	if !clusterActive(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	} else {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
	}