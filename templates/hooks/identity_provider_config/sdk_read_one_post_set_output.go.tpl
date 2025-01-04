	if resp.IdentityProviderConfig.Oidc != nil {
		ko.Spec.Tags = FromACKTags(resp.IdentityProviderConfig.Oidc.Tags)
	}
	temp := string(resp.IdentityProviderConfig.Oidc.Status)
	ko.Status.Status = &temp
	if !identityProviderActive(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	} else {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
	}

