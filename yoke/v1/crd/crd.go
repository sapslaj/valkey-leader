package crd

import (
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	APIGroup   = "valkey-leader.sapslaj.cloud"
	APIVersion = "valkey-leader.sapslaj.cloud/v1"
	KindValkey = "Valkey"
)

type Valkey struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              ValkeySpec `json:"spec"`
}

type ValkeyLeader struct {
	Enabled         *bool           `json:"enabled,omitempty"`
	LeaderLeaseName string          `json:"leaderLeaseName"`
	TargetService   string          `json:"targetService"`
	Image           string          `json:"image"`
	ImageTag        string          `json:"imageTag"`
	Env             []corev1.EnvVar `json:"env"`
}

type RedisExporter struct {
	Enabled  *bool           `json:"enabled,omitempty"`
	Image    string          `json:"image"`
	ImageTag string          `json:"imageTag"`
	Env      []corev1.EnvVar `json:"env"`
}

type ServiceConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Name    string `json:"name"`
	Port    int32  `json:"port"`
}

type Services struct {
	Headless  ServiceConfig `json:"headless"`
	Read      ServiceConfig `json:"read"`
	ReadOnly  ServiceConfig `json:"readOnly"`
	ReadWrite ServiceConfig `json:"readWrite"`
	Metrics   ServiceConfig `json:"metrics"`
}

type ServiceAccount struct {
	Create      bool              `json:"create"`
	Name        string            `json:"name"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type RBAC struct {
	Create bool `json:"create"`
}

type ValkeySpec struct {
	MinReadySeconds      int                              `json:"minReadySeconds"`
	PodManagementPolicy  string                           `json:"podManagementPolicy"`
	Replicas             *int32                           `json:"replicas,omitempty"`
	LabelSelector        map[string]string                `json:"labelSelector"`
	Template             corev1.PodTemplateSpec           `json:"template"`
	UpdateStrategy       appsv1.StatefulSetUpdateStrategy `json:"updateStrategy"`
	VolumeClaimTemplates []corev1.PersistentVolumeClaim   `json:"volumeClaimTemplates"`
	Image                string                           `json:"image"`
	ImageTag             string                           `json:"imageTag"`
	Services             Services                         `json:"services"`
	ValkeyLeader         ValkeyLeader                     `json:"valkeyLeader"`
	RedisExporter        RedisExporter                    `json:"redisExporter"`
	ServiceAccount       ServiceAccount                   `json:"serviceAccount"`
	RBAC                 RBAC                             `json:"rbac"`
}

func (backend Valkey) MarshalJSON() ([]byte, error) {
	backend.Kind = KindValkey
	backend.APIVersion = APIVersion

	type ValkeyAlt Valkey
	return json.Marshal(ValkeyAlt(backend))
}

func (backend *Valkey) UnmarshalJSON(data []byte) error {
	type ValkeyAlt Valkey
	if err := json.Unmarshal(data, (*ValkeyAlt)(backend)); err != nil {
		return err
	}
	if backend.APIVersion != APIVersion {
		return fmt.Errorf("unexpected api version: expected %s but got %s", APIVersion, backend.APIVersion)
	}
	if backend.Kind != KindValkey {
		return fmt.Errorf("unexpected kind: expected %s but got %s", KindValkey, backend.Kind)
	}
	return nil
}
