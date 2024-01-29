package types

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

type HelmChart struct {
	ApiVersion  string `yaml:"apiVersion"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Version     string `yaml:"version"`
	AppVersion  string `yaml:"appVersion"`
}

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

type Metrics struct {
	Total  int     `json:"total"`
	Detail []*Cast `json:"detail,omitempty" protobuf:"bytes,3,opt,name=detail"`
}
