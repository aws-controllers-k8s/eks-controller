	// After successfully issuing the DeleteCluster API call, the cluster
	// transitions to DELETING state. We must wait until the cluster is fully
	// gone (DescribeCluster returns ResourceNotFoundException) before allowing
	// the finalizer to be removed. This prevents a race condition where the
	// IAM Role CR attempts deletion while EKS-managed instance profiles are
	// still attached to the node role (causing DeleteConflict errors).
	// See: https://github.com/aws-controllers-k8s/iam-controller/pull/181
	if err == nil {
		return r, requeueWaitWhileDeleting
	}
