	if addonDeleting(latest) {
		msg := "Addon is currently being deleted"
		ackcondition.SetSynced(latest, corev1.ConditionFalse, &msg, nil)
		return latest, requeueWaitWhileDeleting
	}

	// Check if addon is in a failed state that requires retry
	inFailedState := addonInFailedState(latest)

	if !addonActive(latest) && !inFailedState {
		msg := "Addon is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(latest, corev1.ConditionFalse, &msg, nil)
		if addonHasTerminalStatus(latest) {
			ackcondition.SetTerminal(latest, corev1.ConditionTrue, &msg, nil)
			return latest, nil
		}
		return latest, requeueWaitUntilCanModify(latest)
	}

	// If addon is in failed state, we need to force an update regardless of delta
	if inFailedState {
		msg := "Addon is in '" + *latest.ko.Status.Status + "' status, attempting recovery"
		ackcondition.SetSynced(latest, corev1.ConditionFalse, &msg, nil)
	}

	if delta.DifferentAt("Spec.Tags") {
		err := syncTags(
			ctx, rm.sdkapi, rm.metrics,
			string(*desired.ko.Status.ACKResourceMetadata.ARN),
			aws.ToStringMap(desired.ko.Spec.Tags), aws.ToStringMap(latest.ko.Spec.Tags),
		)
		if err != nil {
			return nil, err
		}
	}
	// If addon is in failed state, always attempt update to recover
	// Otherwise, check if there are differences to update
	if !inFailedState && !delta.DifferentExcept("Spec.Tags") {
		return desired, nil
	}