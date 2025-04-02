// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
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

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib
// Changes customizing the original source code

package swok8sobjectsreceiver // import "github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swok8sobjectsreceiver"

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/extension/xextension/storage"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	apiWatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/watch"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swok8sobjectsreceiver/internal/metadata"
)

type k8sobjectsreceiver struct {
	setting         receiver.Settings
	config          *Config
	stopperChanList []chan struct{}
	client          dynamic.Interface
	consumer        consumer.Logs
	obsrecv         *receiverhelper.ObsReport
	mu              sync.Mutex
	cancel          context.CancelFunc
	storageClient   storage.Client
}

type objectstorage struct {
	key     string
	Objects map[string]objecthashes `json:"objects"`
	mu      sync.Mutex
}

type objecthashes struct {
	Metadata string `json:"metadata"`
	Spec     string `json:"spec"`
	Status   string `json:"status"`
	Other    string `json:"other"`
}

func newReceiver(params receiver.Settings, config *Config, consumer consumer.Logs) (receiver.Logs, error) {
	transport := "http"

	obsrecv, err := receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             params.ID,
		Transport:              transport,
		ReceiverCreateSettings: params,
	})
	if err != nil {
		return nil, err
	}

	for _, object := range config.Objects {
		object.exclude = make(map[apiWatch.EventType]bool)
		for _, item := range object.ExcludeWatchType {
			object.exclude[item] = true
		}
	}

	return &k8sobjectsreceiver{
		setting:  params,
		consumer: consumer,
		config:   config,
		obsrecv:  obsrecv,
		mu:       sync.Mutex{},
	}, nil
}

func (kr *k8sobjectsreceiver) Start(ctx context.Context, host component.Host) error {
	client, err := kr.config.getDynamicClient()
	if err != nil {
		return err
	}
	kr.client = client
	kr.setting.Logger.Info("Object Receiver started")

	cctx, cancel := context.WithCancel(ctx)
	kr.cancel = cancel

	if kr.config.StorageID != nil {
		kr.storageClient, err = adapter.GetStorageClient(ctx, host, kr.config.StorageID, kr.setting.ID)
		if err != nil {
			return fmt.Errorf("error connecting to storage: %w", err)
		}
	}

	for _, object := range kr.config.Objects {
		kr.start(cctx, object)
	}
	return nil
}

func (kr *k8sobjectsreceiver) Shutdown(ctx context.Context) error {
	kr.setting.Logger.Info("Object Receiver stopped")
	if kr.cancel != nil {
		kr.cancel()
	}

	kr.mu.Lock()
	for _, stopperChan := range kr.stopperChanList {
		close(stopperChan)
	}
	if kr.storageClient != nil {
		kr.storageClient.Close(ctx)
		kr.storageClient = nil
	}
	kr.mu.Unlock()
	return nil
}

func (kr *k8sobjectsreceiver) start(ctx context.Context, object *K8sObjectsConfig) {
	resource := kr.client.Resource(*object.gvr)
	kr.setting.Logger.Info("Started collecting", zap.Any("gvr", object.gvr), zap.Any("mode", object.Mode), zap.Any("namespaces", object.Namespaces))

	switch object.Mode {
	case PullMode:
		if len(object.Namespaces) == 0 {
			go kr.startPull(ctx, object, resource)
		} else {
			for _, ns := range object.Namespaces {
				go kr.startPull(ctx, object, resource.Namespace(ns))
			}
		}

	case WatchMode:
		if len(object.Namespaces) == 0 {
			go kr.startWatch(ctx, object, resource)
		} else {
			for _, ns := range object.Namespaces {
				go kr.startWatch(ctx, object, resource.Namespace(ns))
			}
		}
	}
}

func (kr *k8sobjectsreceiver) startPull(ctx context.Context, config *K8sObjectsConfig, resource dynamic.ResourceInterface) {
	stopperChan := make(chan struct{})
	kr.mu.Lock()
	kr.stopperChanList = append(kr.stopperChanList, stopperChan)
	kr.mu.Unlock()
	ticker := newTicker(ctx, config.Interval)
	listOption := metav1.ListOptions{
		FieldSelector: config.FieldSelector,
		LabelSelector: config.LabelSelector,
	}

	if config.ResourceVersion != "" {
		listOption.ResourceVersion = config.ResourceVersion
		listOption.ResourceVersionMatch = metav1.ResourceVersionMatchExact
	}

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			objects, err := resource.List(ctx, listOption)
			if err != nil {
				kr.setting.Logger.Error("error in pulling object", zap.String("resource", config.gvr.String()), zap.Error(err))
			} else if len(objects.Items) > 0 {
				// clear out managed fields
				for _, item := range objects.Items {
					item.SetManagedFields(nil)
				}
				logs := pullObjectsToLogData(objects, time.Now(), config)
				obsCtx := kr.obsrecv.StartLogsOp(ctx)
				logRecordCount := logs.LogRecordCount()
				err = kr.consumer.ConsumeLogs(obsCtx, logs)
				kr.obsrecv.EndLogsOp(obsCtx, metadata.Type.String(), logRecordCount, err)
			}
		case <-stopperChan:
			return
		}
	}
}

func (kr *k8sobjectsreceiver) startWatch(ctx context.Context, config *K8sObjectsConfig, resource dynamic.ResourceInterface) {
	stopperChan := make(chan struct{})
	kr.mu.Lock()
	kr.stopperChanList = append(kr.stopperChanList, stopperChan)
	kr.mu.Unlock()

	watchFunc := func(options metav1.ListOptions) (apiWatch.Interface, error) {
		options.FieldSelector = config.FieldSelector
		options.LabelSelector = config.LabelSelector
		return resource.Watch(ctx, options)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	cfgCopy := *config
	storage := &objectstorage{key: config.Name, Objects: map[string]objecthashes{}}

	wait.UntilWithContext(cancelCtx, func(newCtx context.Context) {
		resourceVersion, err := kr.getResourceVersionAndUpdateCache(newCtx, &cfgCopy, resource, storage)
		if err != nil {
			kr.setting.Logger.Error("could not retrieve a resourceVersion", zap.String("resource", cfgCopy.gvr.String()), zap.Error(err))
			cancel()
			return
		}

		done := kr.doWatch(newCtx, &cfgCopy, resourceVersion, watchFunc, stopperChan, storage)
		if done {
			cancel()
			return
		}

		// need to restart with a fresh resource version
		cfgCopy.ResourceVersion = ""
	}, 0)
}

// doWatch returns true when watching is done, false when watching should be restarted.
func (kr *k8sobjectsreceiver) doWatch(ctx context.Context, config *K8sObjectsConfig, resourceVersion string, watchFunc func(options metav1.ListOptions) (apiWatch.Interface, error), stopperChan chan struct{}, storage *objectstorage) bool {
	watcher, err := watch.NewRetryWatcher(resourceVersion, &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		kr.setting.Logger.Error("error in watching object", zap.String("resource", config.gvr.String()), zap.Error(err))
		return true
	}

	defer watcher.Stop()
	res := watcher.ResultChan()
	for {
		select {
		case data, ok := <-res:
			if data.Type == apiWatch.Error {
				errObject := apierrors.FromObject(data.Object)
				//nolint:errorlint
				if errObject.(*apierrors.StatusError).ErrStatus.Code == http.StatusGone {
					kr.setting.Logger.Info("received a 410, grabbing new resource version", zap.Any("data", data))
					// we received a 410 so we need to restart
					return false
				}
			}

			if !ok {
				kr.setting.Logger.Warn("Watch channel closed unexpectedly", zap.String("resource", config.gvr.String()))
				return true
			}

			if config.exclude[data.Type] {
				kr.setting.Logger.Debug("dropping excluded data", zap.String("type", string(data.Type)))
				continue
			}

			err = kr.watchEventToLogData(ctx, &data, time.Now(), config, storage, false)
			if err != nil {
				kr.setting.Logger.Error("error converting objects to log data", zap.Error(err))
			}

			storage.mu.Lock()
			err = kr.saveStorage(ctx, storage)
			if err != nil {
				kr.setting.Logger.Error("error saving storage", zap.Error(err))
			}
			storage.mu.Unlock()

		case <-stopperChan:
			watcher.Stop()
			return true
		}
	}
}

func (kr *k8sobjectsreceiver) watchEventToLogData(ctx context.Context, event *apiWatch.Event, observedAt time.Time, config *K8sObjectsConfig, storage *objectstorage, initialPoll bool) error {
	udata, ok := event.Object.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("received data that wasnt unstructure, %v", event)
	}

	// clear out managed fields
	udata.SetManagedFields(nil)

	key := getKey(udata)
	hashes, err := getObjectHashes(udata)
	if err != nil {
		return fmt.Errorf("not able to get hashes from object %s error: %w", key, err)
	}

	statusChanged := true
	metadataChanged := true
	specChanged := true
	otherChanged := true

	storage.mu.Lock()
	oldHashes, exists := storage.Objects[key]
	if exists {
		metadataChanged = oldHashes.Metadata != hashes.Metadata
		specChanged = oldHashes.Spec != hashes.Spec
		statusChanged = oldHashes.Status != hashes.Status
		otherChanged = oldHashes.Other != hashes.Other
	}
	storage.Objects[key] = *hashes

	if event.Type == apiWatch.Deleted {
		delete(storage.Objects, key)
	}
	storage.mu.Unlock()

	if !metadataChanged && !statusChanged && !specChanged && !otherChanged {
		return nil
	}

	logs, err := watchObjectsToLogData(event, observedAt, config, func(attrs pcommon.Map) {
		attrs.PutBool("sw.initial.poll", initialPoll)
		attrs.PutBool("sw.metadata.changed", metadataChanged)
		attrs.PutBool("sw.status.changed", statusChanged)
		attrs.PutBool("sw.spec.changed", specChanged)
		attrs.PutBool("sw.other.changed", otherChanged)
	})
	if err != nil {
		return fmt.Errorf("error converting watch objects to log data %w", err)
	} else {
		obsCtx := kr.obsrecv.StartLogsOp(ctx)
		err := kr.consumer.ConsumeLogs(obsCtx, logs)
		kr.obsrecv.EndLogsOp(obsCtx, metadata.Type.String(), 1, err)
	}

	return nil
}

func (kr *k8sobjectsreceiver) saveStorage(ctx context.Context, storage *objectstorage) error {
	if kr.storageClient == nil || storage == nil {
		return nil
	}

	dataBytes, err := json.Marshal(storage.Objects)
	if err != nil {
		return err
	}

	return kr.storageClient.Set(ctx, storage.key, dataBytes)
}

func (kr *k8sobjectsreceiver) loadStorage(ctx context.Context, storage *objectstorage) error {
	if kr.storageClient == nil || storage == nil {
		return nil
	}

	storage.mu.Lock()
	defer storage.mu.Unlock()

	// load existing data from storage
	dataBytes, err := kr.storageClient.Get(ctx, storage.key)
	if err != nil {
		return err
	}

	if len(dataBytes) > 0 {
		err = json.Unmarshal(dataBytes, &storage.Objects)
		if err != nil {
			kr.setting.Logger.Error("failed to unmarshal stored data", zap.Error(err), zap.Any("key", storage.key))

			// clear the storage if we can't unmarshal the data
			storage.Objects = map[string]objecthashes{}
			return nil
		}

		kr.setting.Logger.Info("Data loaded from storage", zap.Any("key", storage.key))
	}

	return nil
}

func getObjectHashes(udata *unstructured.Unstructured) (*objecthashes, error) {
	udataCopy := udata.DeepCopy()
	udataCopy.SetResourceVersion("") // do not report changes in resource version

	udataBytes, err := udataCopy.MarshalJSON()
	if err != nil {
		return nil, err
	}

	timestampRegex := regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})\b?`)

	// do not report changes in any timestamps
	cleanedBytes := timestampRegex.ReplaceAll(udataBytes, []byte(""))

	if err = json.Unmarshal(cleanedBytes, &udataCopy.Object); err != nil {
		return nil, err
	}

	metadataHash, err := getHash(udataCopy, "metadata")
	if err != nil {
		return nil, err
	}
	specHash, err := getHash(udataCopy, "spec")
	if err != nil {
		return nil, err
	}
	statusHash, err := getHash(udataCopy, "status")
	if err != nil {
		return nil, err
	}

	unstructured.RemoveNestedField(udataCopy.Object, "metadata")
	unstructured.RemoveNestedField(udataCopy.Object, "spec")
	unstructured.RemoveNestedField(udataCopy.Object, "status")
	otherHash, err := getHash(udataCopy)
	if err != nil {
		return nil, err
	}

	return &objecthashes{
		Metadata: metadataHash,
		Spec:     specHash,
		Status:   statusHash,
		Other:    otherHash,
	}, nil
}

// returns hash of a nested fields or whole object if no fields are provided
func getHash(udata *unstructured.Unstructured, fields ...string) (string, error) {
	if len(fields) == 0 {
		bytes, err := udata.MarshalJSON()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%x", sha256.Sum256(bytes)), nil
	}

	nested, found, err := unstructured.NestedFieldCopy(udata.Object, fields...)
	if err != nil {
		return "", err
	}

	if !found {
		return "", nil
	}

	bytes, err := json.Marshal(nested)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(bytes)), nil
}

// key for the storage
func getKey(udata *unstructured.Unstructured) string {
	return fmt.Sprintf("%s:%s#%s", udata.GetKind(), udata.GetNamespace(), udata.GetName())
}

func (kr *k8sobjectsreceiver) getResourceVersionAndUpdateCache(ctx context.Context, config *K8sObjectsConfig, resource dynamic.ResourceInterface, storage *objectstorage) (string, error) {
	resourceVersion := config.ResourceVersion
	if resourceVersion == "" || resourceVersion == "0" {
		// Proper use of the Kubernetes API Watch capability when no resourceVersion is supplied is to do a list first
		// to get the initial state and a useable resourceVersion.
		// See https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes for details.
		objects, err := resource.List(ctx, metav1.ListOptions{
			FieldSelector: config.FieldSelector,
			LabelSelector: config.LabelSelector,
		})
		if err != nil {
			return "", fmt.Errorf("could not perform initial list for watch on %v, %w", config.gvr.String(), err)
		}
		if objects == nil {
			return "", errors.New("nil objects returned, this is an error in the k8sobjectsreceiver")
		}

		err = kr.loadStorage(ctx, storage)
		if err != nil {
			return "", fmt.Errorf("could not load data from storage, %w", err)
		}

		resourceVersion = objects.GetResourceVersion()
		existingObjects := map[string]struct{}{}

		for _, item := range objects.Items {
			key := getKey(&item)
			existingObjects[key] = struct{}{}

			event := &apiWatch.Event{Type: apiWatch.Added, Object: &item}

			storage.mu.Lock()
			if _, exists := storage.Objects[key]; exists {
				event.Type = apiWatch.Modified
			}
			storage.mu.Unlock()

			err = kr.watchEventToLogData(ctx, event, time.Now(), config, storage, true)
			if err != nil {
				kr.setting.Logger.Error("error converting objects to log data", zap.Error(err))
			}
		}

		//remove objects from storage that no longer exist
		storage.mu.Lock()
		for key := range storage.Objects {
			if _, exists := existingObjects[key]; !exists {
				delete(storage.Objects, key)
			}
		}

		err = kr.saveStorage(ctx, storage)
		if err != nil {
			kr.setting.Logger.Error("error saving storage", zap.Error(err))
		}
		storage.mu.Unlock()

		// If we still don't have a resourceVersion we can try 1 as a last ditch effort.
		// This also helps our unit tests since the fake client can't handle returning resource versions
		// as part of a list of objects.
		if resourceVersion == "" || resourceVersion == "0" {
			resourceVersion = defaultResourceVersion
		}
	}
	return resourceVersion, nil
}

// Start ticking immediately.
// Ref: https://stackoverflow.com/questions/32705582/how-to-get-time-tick-to-tick-immediately
func newTicker(ctx context.Context, repeat time.Duration) *time.Ticker {
	ticker := time.NewTicker(repeat)
	oc := ticker.C
	nc := make(chan time.Time, 1)
	go func() {
		nc <- time.Now()
		for {
			select {
			case tm := <-oc:
				nc <- tm
			case <-ctx.Done():
				return
			}
		}
	}()

	ticker.C = nc
	return ticker
}
