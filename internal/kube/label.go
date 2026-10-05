package kube

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LabelPod adds the job-id label and returns when Kubernetes created the pod.
func LabelPod(ctx context.Context, jobID string) (time.Time, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return time.Time{}, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return time.Time{}, err
	}
	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":{"job-id":%q}}}`, jobID))
	pod, err := cs.CoreV1().Pods(os.Getenv("POD_NAMESPACE")).Patch(
		ctx, os.Getenv("POD_NAME"), types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return time.Time{}, err
	}
	return pod.CreationTimestamp.Time, nil
}