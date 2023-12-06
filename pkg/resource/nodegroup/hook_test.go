package nodegroup

import (
	"testing"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/stretchr/testify/assert"
)

func TestTaints(t *testing.T) {

	noSchedule := "NO_SCHEDULE"
	owner := "owner"
	project := "project"

	teamOne := "teamone"
	projectOne := "projectone"

	a := &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectOne,
					},
				},
			},
		},
	}

	b := &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectOne,
					},
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
				},
			},
		},
	}

	delta := &ackcompare.Delta{}
	customPreCompare(delta, a, b)
	assert.False(t, delta.DifferentAt("Spec.Taints"), "Taints are equals")

	projectTwo := "projecttwo"

	b = &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectTwo,
					},
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
				},
			},
		},
	}

	delta = &ackcompare.Delta{}
	customPreCompare(delta, a, b)
	assert.True(t, delta.DifferentAt("Spec.Taints"), "Taints are different")

	other := "other"

	b = &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectOne,
					},
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
					{
						Effect: &noSchedule,
						Key:    &other,
						Value:  &other,
					},
				},
			},
		},
	}

	delta = &ackcompare.Delta{}
	customPreCompare(delta, a, b)
	assert.True(t, delta.DifferentAt("Spec.Taints"), "Taints have different length")
}
