	if addonDeleting(latest) {
		msg := "Addon is currently being deleted"
		ackcondition.SetSynced(latest, corev1.ConditionFalse, &msg, nil)
		return latest, requeueWaitWhileDeleting
	}
	if !addonActive(latest) {
		msg := "Addon is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(latest, corev1.ConditionFalse, &msg, nil)
		if addonHasTerminalStatus(latest) {
			ackcondition.SetTerminal(latest, corev1.ConditionTrue, &msg, nil)
			return latest, nil
		}
		return latest, requeueWaitUntilCanModify(latest)
	}

	if delta.DifferentAt("Spec.Tags") {
		err := syncTags(
			ctx, rm.sdkapi, rm.metrics, 
			string(*desired.ko.Status.ACKResourceMetadata.ARN), 
			desired.ko.Spec.Tags, latest.ko.Spec.Tags,
		)
		if err != nil {
			return nil, err
		}
	}
    if !delta.DifferentExcept("Spec.Tags"){
        return desired, nil
    }