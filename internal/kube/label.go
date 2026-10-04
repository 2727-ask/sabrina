package kube

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LabelPod adds the job-id label to the pod this solver runs in.
func LabelPod(ctx context.Context, jobID string) error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":{"job-id":%q}}}`, jobID))
	_, err = cs.CoreV1().Pods(os.Getenv("POD_NAMESPACE")).Patch(
		ctx, os.Getenv("POD_NAME"), types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}