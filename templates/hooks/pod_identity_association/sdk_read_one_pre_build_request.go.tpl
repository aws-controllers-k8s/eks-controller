	// Retrieve podIdentityAssociation ID only during adoption 
	if r.ko.Status.AssociationID == nil && runtime.NeedAdoption(r) {
		r.ko.Status.AssociationID, err = rm.getAssociationID(ctx, r)
		if err != nil {
			return nil, err
		}
	}