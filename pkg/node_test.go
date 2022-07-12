package k8status

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func Test_getNodeConditions(t *testing.T) {
	type args struct {
		node v1.Node
	}
	tests := []struct {
		name     string
		node     v1.Node
		ready    bool
		cordoned bool
		messages []string
	}{
		{
			name:     "no node status",
			node:     v1.Node{},
			ready:    false,
			cordoned: false,
			messages: []string{},
		},
		{
			name: "cordoned node",
			node: v1.Node{
				Spec: v1.NodeSpec{
					Unschedulable: true,
				},
			},
			ready:    false,
			cordoned: true,
			messages: []string{},
		},
		{
			name: "ready node",
			node: v1.Node{
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			ready:    true,
			cordoned: false,
			messages: []string{},
		},
		{
			name: "node with disk pressure issue",
			node: v1.Node{
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:    v1.NodeDiskPressure,
							Status:  v1.ConditionTrue,
							Message: "node has disk pressure",
						},
					},
				},
			},
			ready:    true,
			cordoned: false,
			messages: []string{
				"node has disk pressure",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, cordoned, messages := getNodeConditions(tt.node)
			if ready != tt.ready {
				t.Errorf("getNodeConditions() ready = %v, want %v", ready, tt.ready)
			}
			if cordoned != tt.cordoned {
				t.Errorf("getNodeConditions() cordoned = %v, want %v", cordoned, tt.cordoned)
			}
			if !reflect.DeepEqual(messages, tt.messages) {
				t.Errorf("getNodeConditions() messages = %v, want %v", messages, tt.messages)
			}
		})
	}
}
