	if r.ko.Status.AssociationID == nil {
		r.ko.Status.AssociationID, err = rm.getAssociationID(ctx, r)
		if err != nil {
			return nil, err
		}
	}