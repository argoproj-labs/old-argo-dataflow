package v1alpha1

import (
	"encoding/json"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

type StepSpec struct {
	// +kubebuilder:default=default
	Name      string     `json:"name" protobuf:"bytes,6,opt,name=name"`
	Cat       *Cat       `json:"cat,omitempty" protobuf:"bytes,15,opt,name=cat"`
	Container *Container `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
	Handler   *Handler   `json:"handler,omitempty" protobuf:"bytes,7,opt,name=handler"`
	Git       *Git       `json:"git,omitempty" protobuf:"bytes,12,opt,name=git"`
	Filter    Filter     `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter,casttype=Filter"`
	Map       Map        `json:"map,omitempty" protobuf:"bytes,9,opt,name=map,casttype=Map"`
	Group     *Group     `json:"group,omitempty" protobuf:"bytes,11,opt,name=group"`
	Flatten   *Flatten   `json:"flatten,omitempty" protobuf:"bytes,25,opt,name=flatten"`
	Expand    *Expand    `json:"expand,omitempty" protobuf:"bytes,26,opt,name=expand"`

	Replicas *uint32 `json:"replicas,omitempty" protobuf:"varint,23,opt,name=replicas"`
	Scale    *Scale  `json:"scale,omitempty" protobuf:"bytes,24,opt,name=scale"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sources Sources `json:"sources,omitempty" protobuf:"bytes,3,rep,name=sources"`
	// +patchStrategy=merge
	// +patchMergeKey=name
	Sinks []Sink `json:"sinks,omitempty" protobuf:"bytes,4,rep,name=sinks"`
	// +kubebuilder:default=OnFailure
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,5,opt,name=restartPolicy,casttype=k8s.io/api/core/v1.RestartPolicy"`
	Terminator    bool                 `json:"terminator,omitempty" protobuf:"varint,10,opt,name=terminator"` // if this step terminates, terminate all steps in the pipeline
	// +patchStrategy=merge
	// +patchMergeKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty" protobuf:"bytes,13,rep,name=volumes"`
	// +kubebuilder:default=pipeline
	ServiceAccountName string              `json:"serviceAccountName,omitempty" protobuf:"bytes,14,opt,name=serviceAccountName"`
	Metadata           *Metadata           `json:"metadata,omitempty" protobuf:"bytes,16,opt,name=metadata"`
	NodeSelector       map[string]string   `json:"nodeSelector,omitempty" protobuf:"bytes,17,rep,name=nodeSelector"`
	Affinity           *corev1.Affinity    `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	Tolerations        []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,19,rep,name=tolerations"`
}

type GetPodSpecReq struct {
	PipelineName   string            `protobuf:"bytes,1,opt,name=pipelineName"`
	Namespace      string            `protobuf:"bytes,2,opt,name=namespace"`
	Replica        int32             `protobuf:"varint,3,opt,name=replica"`
	ImageFormat    string            `protobuf:"bytes,4,opt,name=imageFormat"`
	RunnerImage    string            `protobuf:"bytes,5,opt,name=runnerImage"`
	PullPolicy     corev1.PullPolicy `protobuf:"bytes,6,opt,name=pullPolicy,casttype=k8s.io/api/core/v1.PullPolicy"`
	UpdateInterval time.Duration     `protobuf:"varint,7,opt,name=updateInterval,casttype=time.Duration"`
	StepStatus     StepStatus        `protobuf:"bytes,8,opt,name=stepStatus"`
	BearerToken    string            `protobuf:"bytes,9,opt,name=bearerToken"`
}

func (in *StepSpec) GetPodSpec(req GetPodSpecReq) corev1.PodSpec {
	volume := corev1.Volume{
		Name:         "var-run-argo-dataflow",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
	stepSpec, _ := json.Marshal(in)
	stepStatus, _ := json.Marshal(req.StepStatus)
	volumeMounts := []corev1.VolumeMount{{Name: volume.Name, MountPath: PathVarRun}}

	envVars := []corev1.EnvVar{
		{Name: EnvPipelineName, Value: req.PipelineName},
		{Name: EnvNamespace, Value: req.Namespace},
		{Name: EnvReplica, Value: strconv.Itoa(int(req.Replica))},
		{Name: EnvStepSpec, Value: string(stepSpec)},
		{Name: EnvStepStatus, Value: string(stepStatus)},
		{Name: EnvUpdateInterval, Value: req.UpdateInterval.String()},
		{Name: EnvDataflowBearerToken, Value: req.BearerToken},
	}
	return corev1.PodSpec{
		Volumes:            append(in.Volumes, volume),
		RestartPolicy:      in.RestartPolicy,
		NodeSelector:       in.NodeSelector,
		ServiceAccountName: in.ServiceAccountName,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: pointer.BoolPtr(true),
			RunAsUser:    pointer.Int64Ptr(9653),
		},
		Affinity:    in.Affinity,
		Tolerations: in.Tolerations,
		InitContainers: []corev1.Container{
			{
				Name:            CtrInit,
				Image:           req.RunnerImage,
				ImagePullPolicy: req.PullPolicy,
				Args:            []string{"init"},
				Env:             envVars,
				VolumeMounts:    volumeMounts,
				Resources:       SmallResourceRequirements,
			},
		},
		Containers: []corev1.Container{
			{
				Name:            CtrSidecar,
				Image:           req.RunnerImage,
				ImagePullPolicy: req.PullPolicy,
				Args:            []string{"sidecar"},
				Env:             envVars,
				VolumeMounts:    volumeMounts,
				Resources:       SmallResourceRequirements,
				Ports: []corev1.ContainerPort{
					{ContainerPort: 3569},
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/pre-stop",
							Port: intstr.FromInt(3569),
						},
					},
				},
			},
			in.GetContainer(
				req.ImageFormat,
				req.RunnerImage,
				req.PullPolicy,
				corev1.VolumeMount{Name: "var-run-argo-dataflow", MountPath: "/var/run/argo-dataflow"},
			),
		},
	}
}

func (in *StepSpec) GetIn() *Interface {
	if in.Container != nil {
		return in.Container.GetIn()
	}
	return DefaultInterface
}

func (in *StepSpec) GetContainer(imageFormat, runnerImage string, policy corev1.PullPolicy, mnt corev1.VolumeMount) corev1.Container {
	return in.getType().getContainer(getContainerReq{
		imageFormat:     imageFormat,
		runnerImage:     runnerImage,
		imagePullPolicy: policy,
		volumeMount:     mnt,
		lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{PathPreStop},
				},
			},
		},
	})
}

func (in *StepSpec) getType() containerSupplier {
	if x := in.Cat; x != nil {
		return x
	} else if x := in.Container; x != nil {
		return x
	} else if x := in.Expand; x != nil {
		return x
	} else if x := in.Filter; x != "" {
		return x
	} else if x := in.Flatten; x != nil {
		return x
	} else if x := in.Git; x != nil {
		return x
	} else if x := in.Group; x != nil {
		return x
	} else if x := in.Handler; x != nil {
		return x
	} else if x := in.Map; x != "" {
		return x
	} else {
		panic("invalid step spec")
	}
}

func (in *StepSpec) CalculateReplicas(pending int) int {
	if in.Replicas != nil {
		return int(*in.Replicas)
	}
	if in.Scale == nil {
		return 1
	}
	return in.Scale.Calculate(pending)
}
