	if nodegroupDeleting(r) {
		return r, requeueWaitWhileDeleting
	}
