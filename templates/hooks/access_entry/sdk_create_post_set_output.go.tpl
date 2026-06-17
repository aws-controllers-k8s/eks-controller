        if desired.ko.Spec.AccessPolicies != nil {
                ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
                return &resource{ko}, nil
        }
