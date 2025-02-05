	if delta.DifferentAt("Spec.Tags") {
		// TODO(a-hilaly) we need to switch to "ONLY" using the ARN from the ackResourceMetadata
		// in the future.
		resourceARN := ""
		if desired.ko.Status.ACKResourceMetadata.ARN != nil {
			resourceARN = *latest.ko.Status.AssociationARN
		} else if desired.ko.Status.AssociationARN != nil{
			resourceARN = string(*latest.ko.Status.ACKResourceMetadata.ARN)
		}
		err := syncTags(
			ctx, rm.sdkapi, rm.metrics,
			resourceARN,
			ToACKTags(desired.ko.Spec.Tags), ToACKTags(latest.ko.Spec.Tags),
		)
		if err != nil {
			return nil, err
		}
	}
    if !delta.DifferentExcept("Spec.Tags"){
        return desired, nil
    }