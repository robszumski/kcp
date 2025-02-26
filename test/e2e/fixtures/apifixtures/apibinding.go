/*
Copyright 2022 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apifixtures

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kcp-dev/logicalcluster"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
)

// BindToExport creates an APIBinding in bindingClusterName that points at apiExportName in exportClusterName. It waits
// up to wait.ForeverTestTimeout for the APIBinding to have its apisv1alpha1.InitialBindingCompleted status.
func BindToExport(
	ctx context.Context,
	t *testing.T,
	exportClusterName logicalcluster.Name,
	apiExportName string,
	bindingClusterName logicalcluster.Name,
	clusterClient kcpclientset.ClusterInterface,
) {
	binding := &apisv1alpha1.APIBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.Replace(exportClusterName.String(), ":", "-", -1),
		},
		Spec: apisv1alpha1.APIBindingSpec{
			Reference: apisv1alpha1.ExportReference{
				Workspace: &apisv1alpha1.WorkspaceExportReference{
					Path:       exportClusterName.String(),
					ExportName: apiExportName,
				},
			},
		},
	}

	t.Logf("Creating APIBinding %s|%s", bindingClusterName, binding.Name)
	_, err := clusterClient.Cluster(bindingClusterName).ApisV1alpha1().APIBindings().Create(ctx, binding, metav1.CreateOptions{})
	require.NoError(t, err, "error creating APIBinding %s|%s", bindingClusterName, binding.Name)

	require.Eventually(t, func() bool {
		b, err := clusterClient.Cluster(bindingClusterName).ApisV1alpha1().APIBindings().Get(ctx, binding.Name, metav1.GetOptions{})
		if err != nil {
			t.Logf("Unexpected error getting APIBinding %s|%s: %v", bindingClusterName, binding.Name, err)
			return false
		}

		return conditions.IsTrue(b, apisv1alpha1.InitialBindingCompleted)
	}, wait.ForeverTestTimeout, 100*time.Millisecond)
}
