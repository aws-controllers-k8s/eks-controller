	if clusterDeleting(r) {
		return r, requeueWaitWhileDeleting
	}
	inUse, err := rm.clusterInUse(ctx, r);
	if err != nil {
		return nil, err
	} else if inUse {
		return r, requeueWaitWhileInUse
	}
