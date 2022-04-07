package flags

import (
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// BuildRunSpecFromFlags creates a BuildRun spec from command-line flags.
func BuildRunSpecFromFlags(flags *pflag.FlagSet) *buildv1alpha1.BuildRunSpec {
	spec := &buildv1alpha1.BuildRunSpec{
		BuildRef: &buildv1alpha1.BuildRef{
			APIVersion: pointer.String(""),
		},
		ServiceAccount: &buildv1alpha1.ServiceAccount{
			Name:     pointer.String(""),
			Generate: pointer.Bool(false),
		},
		Timeout: &metav1.Duration{},
		Output: &buildv1alpha1.Image{
			Credentials: &corev1.LocalObjectReference{},
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Env: []corev1.EnvVar{},
	}

	buildRefFlags(flags, spec.BuildRef)
	serviceAccountFlags(flags, spec.ServiceAccount)
	timeoutFlags(flags, spec.Timeout)
	imageFlags(flags, "output", spec.Output)
	envFlags(flags, &spec.Env)
	imageLabelsFlags(flags, spec.Output.Labels)
	imageAnnotationsFlags(flags, spec.Output.Annotations)

	return spec
}

// SanitizeBuildRunSpec checks for empty inner data structures and replaces them with nil.
func SanitizeBuildRunSpec(br *buildv1alpha1.BuildRunSpec) {
	if br == nil {
		return
	}
	if br.BuildRef != nil {
		if br.BuildRef.Name == "" && br.BuildRef.APIVersion != nil && *br.BuildRef.APIVersion == "" {
			br.BuildRef = nil
		}
	}
	if br.ServiceAccount != nil {
		if (br.ServiceAccount.Name == nil || *br.ServiceAccount.Name == "") &&
			(br.ServiceAccount.Generate == nil || !*br.ServiceAccount.Generate) {
			br.ServiceAccount = nil
		}
	}
	if br.Output != nil {
		if br.Output.Credentials != nil && br.Output.Credentials.Name == "" {
			br.Output.Credentials = nil
		}
		if br.Output.Image == "" && br.Output.Credentials == nil {
			br.Output = nil
		}
	}
	if br.Timeout != nil && br.Timeout.Duration == 0 {
		br.Timeout = nil
	}

	if len(br.Env) == 0 {
		br.Env = nil
	}
}
