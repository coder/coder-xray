package reporter_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"cdr.dev/slog/sloggers/slogtest"

	"github.com/coder/coder-xray/jfrog"
	"github.com/coder/coder-xray/reporter"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/codersdk/agentsdk"
)

func TestK8SReporter(t *testing.T) {
	t.Parallel()

	const (
		expectedCrit       = 12
		expectedHigh       = 10
		expectedMedium     = 5
		expectedImage      = "docker.io/my-repo/ubuntu:22.04"
		expectedNamespace  = "test-namespace"
		expectedCoderToken = "abc123"
		expectedAgentToken = "test-token"
	)

	var (
		ctx = context.Background()

		expectedAgentID     = uuid.New()
		expectedWorkspaceID = uuid.New()

		k8sClient   = fake.NewSimpleClientset()
		coderClient = reporter.NewMockCoderClient(gomock.NewController(t))
		jfrogClient = jfrog.NewMockClient(gomock.NewController(t))
		resultsCh   = make(chan codersdk.JFrogXrayScan)
	)

	img := jfrog.Image{
		Repo:    "my-repo",
		Package: "ubuntu",
		Version: "22.04",
	}

	xrayResult := jfrog.ScanResult{
		Version: "22.04",
		SecurityIssues: jfrog.SecurityIssues{
			Critical: expectedCrit,
			High:     expectedHigh,
			Medium:   expectedMedium,
			Total:    expectedCrit + expectedHigh + expectedMedium,
		},
		PackageID: "docker://my-repo/ubuntu",
	}

	jfrogClient.EXPECT().ScanResults(img).Return(xrayResult, nil)

	jfrogClient.EXPECT().ResultsURL(img, xrayResult.PackageID)

	coderClient.EXPECT().AgentManifest(ctx, expectedAgentToken).Return(agentsdk.Manifest{
		WorkspaceID: expectedWorkspaceID,
		AgentID:     expectedAgentID,
	}, nil)

	coderClient.EXPECT().PostJFrogXrayScan(ctx, codersdk.JFrogXrayScan{
		WorkspaceID: expectedWorkspaceID,
		AgentID:     expectedAgentID,
		Critical:    expectedCrit,
		High:        expectedHigh,
		Medium:      expectedMedium,
	})

	rep := reporter.K8sReporter{
		Client:      k8sClient,
		Namespace:   expectedNamespace,
		Logger:      slogtest.Make(t, nil),
		CoderClient: coderClient,
		JFrogClient: jfrogClient,
		ResultsChan: resultsCh,
	}

	err := rep.Init(ctx)
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

	rep.Start(nil)

	expectedResult := codersdk.JFrogXrayScan{
		WorkspaceID: expectedWorkspaceID,
		AgentID:     expectedAgentID,
		Critical:    expectedCrit,
		High:        expectedHigh,
		Medium:      expectedMedium,
	}
	select {
	case actualResult := <-resultsCh:
		require.Equal(t, expectedResult, actualResult)
	case <-time.After(time.Second * 10):
		t.Fatalf("ctx timed out waiting for result")
	}
}
