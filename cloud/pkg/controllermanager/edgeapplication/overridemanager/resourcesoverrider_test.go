/*
Copyright 2024 The KubeEdge Authors.

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

package overridemanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeedge/api/apis/apps/v1alpha1"
)

func TestResourcesOverrider_ApplyOverrides(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name          string
		rawObj        *unstructured.Unstructured
		overriders    OverriderInfo
		expectedObj   map[string]interface{}
		expectedError bool
	}{
		{
			name: "Override Pod resources",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "container1",
								"resources": map[string]interface{}{
									"limits": map[string]interface{}{
										"cpu":    "100m",
										"memory": "128Mi",
									},
								},
							},
						},
					},
				},
			},
			overriders: OverriderInfo{
				Overriders: &v1alpha1.Overriders{
					ResourcesOverriders: []v1alpha1.ResourcesOverrider{
						{
							ContainerName: "container1",
							Value: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
			expectedObj: map[string]interface{}{
				"kind": "Pod",
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name": "container1",
							"resources": map[string]interface{}{
								"limits": map[string]interface{}{
									"cpu":    "200m",
									"memory": "256Mi",
								},
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Override Deployment resources",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name": "container1",
										"resources": map[string]interface{}{
											"limits": map[string]interface{}{
												"cpu":    "100m",
												"memory": "128Mi",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			overriders: OverriderInfo{
				Overriders: &v1alpha1.Overriders{
					ResourcesOverriders: []v1alpha1.ResourcesOverrider{
						{
							ContainerName: "container1",
							Value: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
			expectedObj: map[string]interface{}{
				"kind": "Deployment",
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name": "container1",
									"resources": map[string]interface{}{
										"limits": map[string]interface{}{
											"cpu":    "200m",
											"memory": "256Mi",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "No matching container",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "container1",
							},
						},
					},
				},
			},
			overriders: OverriderInfo{
				Overriders: &v1alpha1.Overriders{
					ResourcesOverriders: []v1alpha1.ResourcesOverrider{
						{
							ContainerName: "container2",
							Value: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
			expectedObj: map[string]interface{}{
				"kind": "Pod",
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name": "container1",
						},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrider := &ResourcesOverrider{}
			err := overrider.ApplyOverrides(tt.rawObj, tt.overriders)

			if tt.expectedError {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedObj, tt.rawObj.Object)
			}
		})
	}
}

func TestBuildResourcesPatches(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name               string
		rawObj             *unstructured.Unstructured
		resourcesOverrider *v1alpha1.ResourcesOverrider
		expectedPatches    []overrideOption
		expectedError      bool
	}{
		{
			name: "Build patches for Pod",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "container1",
							},
						},
					},
				},
			},
			resourcesOverrider: &v1alpha1.ResourcesOverrider{
				ContainerName: "container1",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedPatches: []overrideOption{
				{
					Op:   "replace",
					Path: "/spec/containers/0/resources",
					Value: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Build patches for Deployment",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name": "container1",
									},
								},
							},
						},
					},
				},
			},
			resourcesOverrider: &v1alpha1.ResourcesOverrider{
				ContainerName: "container1",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedPatches: []overrideOption{
				{
					Op:   "replace",
					Path: "/spec/template/spec/containers/0/resources",
					Value: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "No matching container",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "container1",
							},
						},
					},
				},
			},
			resourcesOverrider: &v1alpha1.ResourcesOverrider{
				ContainerName: "container2",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedPatches: []overrideOption{},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches, err := buildResourcesPatches(tt.rawObj, tt.resourcesOverrider)

			if tt.expectedError {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedPatches, patches)
			}
		})
	}
}

func TestBuildResourcesPatchesWithPath(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name               string
		specContainersPath string
		rawObj             *unstructured.Unstructured
		resourcesOverrider *v1alpha1.ResourcesOverrider
		expectedPatches    []overrideOption
		expectedError      bool
	}{
		{
			name:               "Build patches for Pod",
			specContainersPath: "spec/containers",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "container1",
							},
							map[string]interface{}{
								"name": "container2",
							},
						},
					},
				},
			},
			resourcesOverrider: &v1alpha1.ResourcesOverrider{
				ContainerName: "container2",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedPatches: []overrideOption{
				{
					Op:   "replace",
					Path: "/spec/containers/1/resources",
					Value: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:               "Build patches for Deployment",
			specContainersPath: "spec/template/spec/containers",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name": "container1",
									},
									map[string]interface{}{
										"name": "container2",
									},
								},
							},
						},
					},
				},
			},
			resourcesOverrider: &v1alpha1.ResourcesOverrider{
				ContainerName: "container1",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("300m"),
						corev1.ResourceMemory: resource.MustParse("512Mi"),
					},
				},
			},
			expectedPatches: []overrideOption{
				{
					Op:   "replace",
					Path: "/spec/template/spec/containers/0/resources",
					Value: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("300m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:               "No matching container",
			specContainersPath: "spec/containers",
			rawObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "container1",
							},
						},
					},
				},
			},
			resourcesOverrider: &v1alpha1.ResourcesOverrider{
				ContainerName: "container2",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedPatches: []overrideOption{},
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches, err := buildResourcesPatchesWithPath(tt.specContainersPath, tt.rawObj, tt.resourcesOverrider)

			if tt.expectedError {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedPatches, patches)
			}
		})
	}
}

func TestAcquireOverride(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name           string
		resourcesPath  string
		resourcesValue corev1.ResourceRequirements
		expectedOption overrideOption
		expectedError  bool
	}{
		{
			name:          "Valid path and resources",
			resourcesPath: "/spec/containers/0/resources",
			resourcesValue: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			expectedOption: overrideOption{
				Op:   string(v1alpha1.OverriderOpReplace),
				Path: "/spec/containers/0/resources",
				Value: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			expectedError: false,
		},
		{
			name:          "Invalid path (no leading slash)",
			resourcesPath: "spec/containers/0/resources",
			resourcesValue: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			expectedOption: overrideOption{},
			expectedError:  true,
		},
		{
			name:          "Empty resources",
			resourcesPath: "/spec/containers/0/resources",
			resourcesValue: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{},
			},
			expectedOption: overrideOption{
				Op:    string(v1alpha1.OverriderOpReplace),
				Path:  "/spec/containers/0/resources",
				Value: corev1.ResourceRequirements{Limits: corev1.ResourceList{}},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := acquireOverride(tt.resourcesPath, tt.resourcesValue)

			if tt.expectedError {
				assert.Error(err)
				assert.Equal(tt.expectedOption, option)
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedOption, option)
			}
		})
	}
}