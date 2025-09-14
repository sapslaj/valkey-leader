package valkey

import (
	"fmt"
	"maps"

	"github.com/yokecd/yoke/pkg/flight"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sapslaj/valkey-leader/pkg/ptr"
	"github.com/sapslaj/valkey-leader/yoke/v1/crd"
)

const (
	LabelCluster      = "valkey.sapslaj.cloud/cluster"
	LabelServiceType  = "valkey-leader.sapslaj.cloud/service-type"
	LabelInstanceRole = "valkey.sapslaj.cloud/instance-role"

	DefaultRedisPortName   = "redis"
	DefaultRedisPort       = 6379
	DefaultMetricsPortName = "metrics"
	DefaultMetricsPort     = 9121

	DefaultValkeyContainerName        = "valkey"
	DefaultValkeyLeaderContainerName  = "valkey-leader"
	DefaultRedisExporterContainerName = "redis-exporter"

	DefaultValkeyImage        = "valkey/valkey"
	DefaultValkeyTag          = "8"
	DefaultValkeyLeaderImage  = "ghcr.io/sapslaj/valkey-leader"
	DefaultValkeyLeaderTag    = "v0.1.0"
	DefaultRedisExporterImage = "oliver006/redis_exporter"
	DefaultRedisExporterTag   = "v1.66.0"
)

func ImageWithTag(image string, tag string, defaultImage string, defaultTag string) string {
	if image == "" {
		image = defaultImage
	}
	if tag == "" {
		tag = defaultTag
	}
	return image + ":" + tag
}

func ServiceAccountName(valkey crd.Valkey) string {
	name := valkey.ObjectMeta.Name
	if valkey.Spec.ServiceAccount.Name != "" {
		name = valkey.Spec.ServiceAccount.Name
	}
	return name
}

func CreateServiceAccount(valkey crd.Valkey) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ServiceAccountName(valkey),
			Namespace:   valkey.ObjectMeta.Namespace,
			Labels:      valkey.Spec.ServiceAccount.Labels,
			Annotations: valkey.Spec.ServiceAccount.Annotations,
		},
	}
}

func LabelSelector(valkey crd.Valkey, merge ...map[string]string) map[string]string {
	labels := map[string]string{
		LabelCluster: valkey.Name,
	}
	for _, m := range merge {
		maps.Copy(labels, m)
	}
	return labels
}

func ServiceName(valkey crd.Valkey, svcType string, serviceConfig crd.ServiceConfig) string {
	if serviceConfig.Name == "" {
		switch svcType {
		case "headless":
			serviceConfig.Name = "headless"
		case "read":
			serviceConfig.Name = "r"
		case "readOnly":
			serviceConfig.Name = "ro"
		case "readWrite":
			serviceConfig.Name = "rw"
		case "metrics":
			serviceConfig.Name = "metrics"
		}
	}
	svcName := valkey.ObjectMeta.Name
	if serviceConfig.Name != "" {
		svcName += "-"
		svcName += serviceConfig.Name
	}
	return svcName
}

func CreateService(valkey crd.Valkey, svcType string, serviceConfig crd.ServiceConfig) *corev1.Service {
	labels := map[string]string{}
	maps.Copy(labels, valkey.ObjectMeta.Labels)
	labels[LabelServiceType] = svcType

	selectors := map[string]string{}
	switch svcType {
	case "readOnly":
		selectors[LabelInstanceRole] = "replica"
	case "readWrite":
		selectors[LabelInstanceRole] = "primary"
	}

	portName := DefaultRedisPortName
	var port int32 = DefaultRedisPort
	if svcType == DefaultMetricsPortName {
		portName = DefaultMetricsPortName
		port = DefaultMetricsPort
	}
	if serviceConfig.Port != 0 {
		port = serviceConfig.Port
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName(valkey, svcType, serviceConfig),
			Namespace: valkey.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: LabelSelector(valkey, selectors),
			Ports: []corev1.ServicePort{
				{
					Name:       portName,
					Port:       int32(port),
					TargetPort: intstr.FromInt32(port),
				},
			},
		},
	}
}

func CreateServices(valkey crd.Valkey) []*corev1.Service {
	services := []*corev1.Service{}
	if ptr.FromDefault(valkey.Spec.Services.Headless.Enabled, true) {
		services = append(services, CreateService(valkey, "headless", valkey.Spec.Services.Headless))
	}
	if ptr.FromDefault(valkey.Spec.Services.Read.Enabled, true) {
		services = append(services, CreateService(valkey, "read", valkey.Spec.Services.Read))
	}
	if ptr.FromDefault(valkey.Spec.Services.ReadOnly.Enabled, true) {
		services = append(services, CreateService(valkey, "readOnly", valkey.Spec.Services.ReadOnly))
	}
	if ptr.FromDefault(valkey.Spec.Services.ReadWrite.Enabled, true) {
		services = append(services, CreateService(valkey, "readWrite", valkey.Spec.Services.ReadWrite))
	}
	if ptr.FromDefault(valkey.Spec.Services.Metrics.Enabled, true) {
		services = append(services, CreateService(valkey, "metrics", valkey.Spec.Services.Metrics))
	}
	return services
}

func CreateRole(valkey crd.Valkey) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.Identifier(),
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      valkey.Name,
			Namespace: valkey.ObjectMeta.Namespace,
			Labels:    valkey.ObjectMeta.Labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"coordination.k8s.io",
				},
				Resources: []string{
					"leases",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"events",
				},
				Verbs: []string{
					"create",
					"patch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"update",
					"patch",
				},
			},
		},
	}
}

func CreateRoleBinding(
	valkey crd.Valkey,
	serviceAccount *corev1.ServiceAccount,
	role *rbacv1.Role,
) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.Identifier(),
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      valkey.Name,
			Namespace: valkey.ObjectMeta.Namespace,
			Labels:    valkey.ObjectMeta.Labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role.TypeMeta.Kind,
			Name:     role.ObjectMeta.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      serviceAccount.TypeMeta.Kind,
				Name:      serviceAccount.ObjectMeta.Name,
				Namespace: serviceAccount.ObjectMeta.Namespace,
			},
		},
	}
}

func CreateValkeyContainer(valkey crd.Valkey) corev1.Container {
	container := corev1.Container{
		Name: DefaultValkeyContainerName,
	}
	for _, c := range valkey.Spec.Template.Spec.Containers {
		if c.Name == DefaultValkeyContainerName {
			container = c
			break
		}
	}
	if container.Image == "" {
		container.Image = ImageWithTag(
			valkey.Spec.Image,
			valkey.Spec.ImageTag,
			DefaultValkeyImage,
			DefaultValkeyTag,
		)
	}
	foundPort := false
	for i, port := range container.Ports {
		if port.Name != DefaultRedisPortName {
			continue
		}
		foundPort = true
		if port.Protocol == "" {
			port.Protocol = "TCP"
		}
		if port.ContainerPort == 0 {
			port.ContainerPort = DefaultRedisPort
		}
		container.Ports[i] = port
	}
	if !foundPort {
		container.Ports = append(container.Ports, corev1.ContainerPort{
			Name:          DefaultRedisPortName,
			ContainerPort: DefaultRedisPort,
			Protocol:      "TCP",
		})
	}
	return container
}

func CreateValkeyLeaderContainer(valkey crd.Valkey) corev1.Container {
	container := corev1.Container{
		Name: DefaultValkeyLeaderContainerName,
	}
	for _, c := range valkey.Spec.Template.Spec.Containers {
		if c.Name == DefaultValkeyLeaderContainerName {
			container = c
			break
		}
	}
	if container.Image == "" {
		container.Image = ImageWithTag(
			valkey.Spec.ValkeyLeader.Image,
			valkey.Spec.ValkeyLeader.ImageTag,
			DefaultValkeyLeaderImage,
			DefaultValkeyLeaderTag,
		)
	}

	leaseName := valkey.Spec.ValkeyLeader.LeaderLeaseName
	if leaseName == "" {
		leaseName = valkey.ObjectMeta.Name
	}
	container.Env = append([]corev1.EnvVar{
		{
			Name:  "CLUSTER_NAME",
			Value: valkey.ObjectMeta.Name,
		},
		{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "SERVICE_NAME",
			Value: ServiceName(valkey, "headless", valkey.Spec.Services.Headless),
		},
		{
			Name:  "LEADER_LEASE_NAME",
			Value: leaseName,
		},
	}, container.Env...)
	return container
}

func CreateRedisExporterContainer(valkey crd.Valkey) corev1.Container {
	container := corev1.Container{
		Name: DefaultRedisExporterContainerName,
	}
	for _, c := range valkey.Spec.Template.Spec.Containers {
		if c.Name == DefaultRedisExporterContainerName {
			container = c
			break
		}
	}
	if container.Image == "" {
		container.Image = ImageWithTag(
			valkey.Spec.RedisExporter.Image,
			valkey.Spec.RedisExporter.ImageTag,
			DefaultRedisExporterImage,
			DefaultRedisExporterTag,
		)
	}
	foundPort := false
	redisPort := DefaultRedisPort
	for i, port := range container.Ports {
		if port.Name == DefaultRedisPortName && port.ContainerPort != 0 {
			redisPort = int(port.ContainerPort)
			continue
		}
		if port.Name != DefaultMetricsPortName {
			continue
		}
		foundPort = true
		if port.Protocol == "" {
			port.Protocol = "TCP"
		}
		if port.ContainerPort == 0 {
			port.ContainerPort = DefaultMetricsPort
		}
		container.Ports[i] = port
	}
	if !foundPort {
		container.Ports = append(container.Ports, corev1.ContainerPort{
			Name:          DefaultMetricsPortName,
			ContainerPort: DefaultMetricsPort,
			Protocol:      "TCP",
		})
	}
	container.Env = append([]corev1.EnvVar{
		{
			Name:  "REDIS_ADDR",
			Value: fmt.Sprintf("redis://localhost:%d", redisPort),
		},
	}, container.Env...)
	return container
}

func CreatePodTemplateSpec(
	valkey crd.Valkey,
	serviceAccount *corev1.ServiceAccount,
) corev1.PodTemplateSpec {
	template := valkey.Spec.Template
	if template.Spec.ServiceAccountName == "" && serviceAccount != nil {
		template.Spec.ServiceAccountName = serviceAccount.ObjectMeta.Name
	}
	if template.ObjectMeta.Labels == nil {
		template.ObjectMeta.Labels = map[string]string{}
	}
	maps.Copy(template.ObjectMeta.Labels, LabelSelector(valkey))
	foundContainer := false
	for i, container := range template.Spec.Containers {
		if container.Name == DefaultValkeyContainerName {
			foundContainer = true
			template.Spec.Containers[i] = CreateValkeyContainer(valkey)
		}
	}
	if !foundContainer {
		template.Spec.Containers = append(template.Spec.Containers, CreateValkeyContainer(valkey))
	}
	if ptr.FromDefault(valkey.Spec.ValkeyLeader.Enabled, true) {
		foundContainer = false
		for i, container := range template.Spec.Containers {
			if container.Name == DefaultValkeyLeaderContainerName {
				foundContainer = true
				template.Spec.Containers[i] = CreateValkeyLeaderContainer(valkey)
			}
		}
		if !foundContainer {
			template.Spec.Containers = append(template.Spec.Containers, CreateValkeyLeaderContainer(valkey))
		}
	}
	if ptr.FromDefault(valkey.Spec.RedisExporter.Enabled, true) {
		foundContainer = false
		for i, container := range template.Spec.Containers {
			if container.Name == DefaultRedisExporterContainerName {
				foundContainer = true
				template.Spec.Containers[i] = CreateRedisExporterContainer(valkey)
			}
		}
		if !foundContainer {
			template.Spec.Containers = append(template.Spec.Containers, CreateRedisExporterContainer(valkey))
		}
	}
	return template
}

func CreateStatefulSet(
	valkey crd.Valkey,
	serviceAccount *corev1.ServiceAccount,
	services []*corev1.Service,
) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      valkey.Name,
			Namespace: valkey.ObjectMeta.Namespace,
			Labels:    valkey.ObjectMeta.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: services[0].ObjectMeta.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: LabelSelector(valkey),
			},
		},
	}

	if valkey.Spec.Replicas != nil {
		sts.Spec.Replicas = valkey.Spec.Replicas
	}

	sts.Spec.Template = CreatePodTemplateSpec(valkey, serviceAccount)

	return sts
}

func Create(valkey crd.Valkey) []flight.Resource {
	resources := []flight.Resource{}

	serviceAccount := CreateServiceAccount(valkey)
	resources = append(resources, serviceAccount)

	role := CreateRole(valkey)
	resources = append(resources, role)

	roleBinding := CreateRoleBinding(valkey, serviceAccount, role)
	resources = append(resources, roleBinding)

	services := CreateServices(valkey)
	for i := range services {
		resources = append(resources, services[i])
	}

	sts := CreateStatefulSet(valkey, serviceAccount, services)
	resources = append(resources, sts)

	return resources
}
