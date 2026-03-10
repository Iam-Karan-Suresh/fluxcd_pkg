/*
Copyright 2025 The Flux authors

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

package client

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/fluxcd/pkg/apis/meta"
)

func TestCanImpersonate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("returns true when no service account is configured", func(t *testing.T) {
		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		imp := NewImpersonator(c)
		if !imp.CanImpersonate(context.TODO()) {
			t.Error("expected CanImpersonate to return true when no service account is set")
		}
	})

	t.Run("returns true when service account exists in local cluster", func(t *testing.T) {
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "test-ns",
			},
		}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sa).Build()
		imp := NewImpersonator(c,
			WithServiceAccount("", "test-sa", "test-ns"),
		)
		if !imp.CanImpersonate(context.TODO()) {
			t.Error("expected CanImpersonate to return true when SA exists in local cluster")
		}
	})

	t.Run("returns false when service account does not exist in local cluster", func(t *testing.T) {
		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		imp := NewImpersonator(c,
			WithServiceAccount("", "missing-sa", "test-ns"),
		)
		if imp.CanImpersonate(context.TODO()) {
			t.Error("expected CanImpersonate to return false when SA does not exist")
		}
	})

	t.Run("returns true when defaultServiceAccount exists", func(t *testing.T) {
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-sa",
				Namespace: "test-ns",
			},
		}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sa).Build()
		imp := NewImpersonator(c,
			WithServiceAccount("default-sa", "", "test-ns"),
		)
		if !imp.CanImpersonate(context.TODO()) {
			t.Error("expected CanImpersonate to return true when default SA exists")
		}
	})

	t.Run("serviceAccountName overrides defaultServiceAccount", func(t *testing.T) {
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "override-sa",
				Namespace: "test-ns",
			},
		}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sa).Build()
		imp := NewImpersonator(c,
			WithServiceAccount("default-sa", "override-sa", "test-ns"),
		)
		if !imp.CanImpersonate(context.TODO()) {
			t.Error("expected CanImpersonate to return true when overriding SA exists")
		}
	})

	t.Run("returns false when kubeConfigRef is set but cannot build remote client", func(t *testing.T) {
		// When kubeConfigRef is set but the secret doesn't exist, building
		// the remote client fails. CanImpersonate should return false.
		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		imp := NewImpersonator(c,
			WithServiceAccount("", "remote-sa", "test-ns"),
			WithKubeConfig(&meta.KubeConfigReference{
				SecretRef: &meta.SecretKeyReference{
					Name: "nonexistent-secret",
				},
			}, KubeConfigOptions{}, "test-ns", nil),
		)
		if imp.CanImpersonate(context.TODO()) {
			t.Error("expected CanImpersonate to return false when remote client cannot be built")
		}
	})
}
