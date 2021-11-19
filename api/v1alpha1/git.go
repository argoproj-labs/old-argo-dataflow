package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type Git struct {
	Image   string   `json:"image" protobuf:"bytes,1,opt,name=image"`
	Command []string `json:"command,omitempty" protobuf:"bytes,6,rep,name=command"`
	URL     string   `json:"url" protobuf:"bytes,2,opt,name=url"`

	// UsernameSecret is the secret selector to the repository username
	UsernameSecret *corev1.SecretKeySelector `json:"usernameSecret,omitempty" protobuf:"bytes,7,opt,name=usernameSecret"`

	// PasswordSecret is the secret selector to the repository password
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty" protobuf:"bytes,8,opt,name=passwordSecret"`

	// InsecureIgnoreHostKey is the bool value for ignoring check for host key
	InsecureIgnoreHostKey bool `json:"insecureIgnoreHostKey,omitempty" protobuf:"bytes,10,opt,name=insecureIgnoreHostKey"`

	// SSHPrivateKeySecret is the secret selector to the repository ssh private key
	SSHPrivateKeySecret *corev1.SecretKeySelector `json:"sshPrivateKeySecret,omitempty" protobuf:"bytes,9,opt,name=sshPrivateKeySecret"`
	// +kubebuilder:default=.
	Path string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
	// +kubebuilder:default=main
	Branch string          `json:"branch,omitempty" protobuf:"bytes,4,opt,name=branch"`
	Env    []corev1.EnvVar `json:"env,omitempty" protobuf:"bytes,5,rep,name=env"`
}

func (in Git) getContainer(req getContainerReq) corev1.Container {
	return containerBuilder{}.
		init(req).
		image(in.Image).
		command(in.Command...).
		appendEnv(in.Env...).
		workingDir(PathWorkingDir).
		build()
}
