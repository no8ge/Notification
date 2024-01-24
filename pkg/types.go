package types

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

type Msg struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Cast struct {
	Name              string       `json:"name"`
	Namespace         string       `json:"namespace"`
	CreationTimestamp time.Time    `protobuf:"-"`
	Status            v1.PodStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type Castchange struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	CreationTimestamp time.Time `protobuf:"-"`
	Changeset         Changeset `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type Changeset struct {
	Before v1.PodStatus `json:"before,omitempty" protobuf:"bytes,3,opt,name=status"`
	After  v1.PodStatus `json:"after,omitempty" protobuf:"bytes,3,opt,name=status"`
}
