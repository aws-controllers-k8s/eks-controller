	// If a user deleted all the PodIdentityAssociations we should send an empty list to the API
	if delta.DifferentAt("Spec.PodIdentityAssociations") && len(desired.ko.Spec.PodIdentityAssociations) == 0 {
		input.SetPodIdentityAssociations([]*svcsdk.AddonPodIdentityAssociations{})
	}