        if desired.ko.Spec.AccessPolicies != nil {
                msg := "Access policy update pending; resource will be requeued in 30 seconds"
                ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, &msg, nil)
        }
