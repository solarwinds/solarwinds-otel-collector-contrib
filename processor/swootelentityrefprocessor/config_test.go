// Copyright 2026 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package swootelentityrefprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_EmptyAction(t *testing.T) {
	cfg := &Config{Action: ""}
	err := cfg.Validate()
	require.Error(t, err)
}

func TestValidate_UnknownAction(t *testing.T) {
	cfg := &Config{Action: "upsert"}
	err := cfg.Validate()
	require.Error(t, err)
}

func TestValidate_Insert(t *testing.T) {
	cfg := &Config{Action: ActionInsert}
	require.NoError(t, cfg.Validate())
}

func TestValidate_RemoveAll(t *testing.T) {
	cfg := &Config{Action: ActionRemoveAll}
	require.NoError(t, cfg.Validate())
}

func TestValidate_EntityRef_EmptyType(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "", IDKeys: []string{"k8s.pod.name"}},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "type must not be empty")
}

func TestValidate_EntityRef_EmptyIDKeys(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: nil},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "id_keys must not be empty")
}

func TestValidate_RemoveAll_WithEntityRefs_IsError(t *testing.T) {
	cfg := &Config{
		Action: ActionRemoveAll,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "entity_refs must be empty")
}

func TestValidate_EntityRef_Valid(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	require.NoError(t, cfg.Validate())
}

func TestValidate_DuplicateType_SameCase(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate type")
}

func TestValidate_DuplicateType_DifferentCase(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
			{Type: "kubernetespod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name"}},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate type")
}

func TestValidate_EntityRef_EmptyIDKeys_EmptySlice(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{}},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "id_keys must not be empty")
}

func TestValidate_EntityRef_DuplicateIDKey(t *testing.T) {
	cfg := &Config{
		Action: ActionInsert,
		EntityRefs: []EntityRefConfig{
			{Type: "KubernetesPod", IDKeys: []string{"sw.k8s.cluster.uid", "k8s.pod.name", "sw.k8s.cluster.uid"}},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate id_key")
}
