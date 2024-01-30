package reporter_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"cdr.dev/slog/sloggers/slogtest"

	"github.com/coder/xray/reporter"
)

func TestK8SReporter(t *testing.T) {
	t.Parallel()

	const (
		expectedCrit       = 12
		expectedHigh       = 10
		expectedMedium     = 5
		expectedImage      = "docker.io/ubuntu/22.04"
		expectedNamespace  = "test-namespace"
		expectedCoderToken = "abc123"
		expectedAgentToken = "test-token"
	)

	k8sClient := fake.NewSimpleClientset()

	ctx := context.Background()
	rep := reporter.K8sReporter{
		Client:      k8sClient,
		Namespace:   expectedNamespace,
		CoderURL:    &url.URL{},
		Logger:      slogtest.Make(t, nil),
		CoderToken:  expectedCoderToken,
		JFrogClient: jfrogClient,
	}

	err = rep.Init(ctx)
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-pod",
			CreationTimestamp: v1.Time{
				Time: time.Now().Add(time.Hour),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name:  "CODER_AGENT_TOKEN",
							Value: expectedAgentToken,
						},
					},
					Image: expectedImage,
				},
			},
		},
	}
	_, err = k8sClient.CoreV1().Pods(expectedNamespace).Create(ctx, pod, v1.CreateOptions{})
	require.NoError(t, err)
}
