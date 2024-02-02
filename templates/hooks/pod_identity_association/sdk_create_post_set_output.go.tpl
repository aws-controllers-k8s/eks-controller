	if resp.Association.AssociationArn != nil {
		ko.Status.ACKResourceMetadata.ARN = (*ackv1alpha1.AWSResourceName)(resp.Association.AssociationArn)
	}