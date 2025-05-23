// Copyright © 2025 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
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

package apiserver

import (
	"github.com/hyperledger/firefly-common/pkg/config"
)

const (
	Enabled               = "enabled"
	DeprecatedMetricsPath = "path"
	MetricsPath           = "metricsPath"
)

func initDeprecatedMetricsConfig(config config.Section) {
	config.AddKnownKey(Enabled, true)
	config.AddKnownKey(DeprecatedMetricsPath, "/metrics")
}

func initMonitoringConfig(config config.Section) {
	config.AddKnownKey(Enabled, false)
	config.AddKnownKey(MetricsPath, "/metrics")
}
