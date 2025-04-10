	if ko.Spec.ScalingConfig != nil && ko.Spec.ScalingConfig.DesiredSize != nil {
		ko.Status.DesiredSize = ko.Spec.ScalingConfig.DesiredSize
	}

	if !nodegroupActive(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	} else {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
	}

