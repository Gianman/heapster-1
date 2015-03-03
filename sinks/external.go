// Copyright 2015 Google Inc. All Rights Reserved.
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

package sinks

import (
	"fmt"

	sink_api "github.com/GoogleCloudPlatform/heapster/sinks/api"
	"github.com/GoogleCloudPlatform/heapster/sinks/gcm"
)

type externalSinkManager struct {
	decoder       sink_api.Decoder
	externalSinks []sink_api.ExternalSink
}

func NewGcmSink() (Sink, error) {
	externalSink, err := gcm.NewGCMExternalSink()
	if err != nil {
		return nil, err
	}

	// Get supported metrics.
	supportedMetrics := sink_api.SupportedStatMetrics()
	for i := range supportedMetrics {
		supportedMetrics[i].Labels = sink_api.SupportedLabels()
	}

	// Create the metrics.
	descriptors := make([]sink_api.MetricDescriptor, 0, len(supportedMetrics))
	for _, supported := range supportedMetrics {
		descriptors = append(descriptors, supported.MetricDescriptor)
	}
	err = externalSink.Register(descriptors)
	if err != nil {
		return nil, err
	}
	decoder := sink_api.NewDecoder()
	return &externalSinkManager{
		externalSinks: []sink_api.ExternalSink{externalSink},
		decoder:       decoder,
	}, nil
}

// TODO(vmarmol): Paralellize this.
func (self *externalSinkManager) StoreData(input Data) error {
	timeseries, err := self.decoder.Metrics(input)
	if err != nil {
		return err
	}
	// Format metrics and push them.
	var errors []error
	for _, externalSink := range self.externalSinks {
		if err := externalSink.Store(timeseries); err != nil {
			errors = append(errors, err)
		}
	}
	err = nil
	if len(errors) > 0 {
		errStr := ""
		for _, err := range errors {
			errStr = fmt.Sprintf("%v ", err)
		}
		err = fmt.Errorf("encountered following errors: %s", errStr)
	}

	return err
}

func (self *externalSinkManager) GetConfig() string {
	desc := "External Sinks\n"

	// Add metrics being exported.
	desc += "\tExported metrics:"
	for _, supported := range sink_api.SupportedStatMetrics() {
		desc += fmt.Sprintf("\t\t%s: %s", supported.Name, supported.Description)
	}

	// Add labels being used.
	desc += "\tExported labels:"
	for _, label := range sink_api.SupportedLabels() {
		desc += fmt.Sprintf("\t\t%s: %s", label.Key, label.Description)
	}
	desc += "\n\tExternal Sinks:"
	for _, externalSink := range self.externalSinks {
		desc += fmt.Sprintf("\n\t\t%s", externalSink.DebugInfo())
	}

	return desc
}
