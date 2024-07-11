	// Updating addons will very likely change the state of the addon
	// so we should requeue the resource to check the status again.
	returnAddonUpdating(&resource{ko})