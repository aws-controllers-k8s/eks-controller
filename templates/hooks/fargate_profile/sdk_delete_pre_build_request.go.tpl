	if profileDeleting(r) {
		return r, requeueWaitWhileDeleting
	}
