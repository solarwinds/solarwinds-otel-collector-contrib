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

package swok8sdiscovery

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	defaultInterval time.Duration = time.Minute * 5
)

type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`

	Interval time.Duration `mapstructure:"interval"`

	Reporter string `mapstructure:"reporter"`

	Database *DatabaseDiscoveryConfig `mapstructure:"database"`

	// For mocking purposes only.
	makeClient func() (k8s.Interface, error)
}

type DatabaseDiscoveryConfig struct {
	ImageRules  []ImageRule  `mapstructure:"image_rules"`
	DomainRules []DomainRule `mapstructure:"domain_rules"`
}

type ImageRule struct {
	DatabaseType string `mapstructure:"database_type"`
	// regular expressions patterns to match against container images
	Patterns         []string         `mapstructure:"patterns"`
	PatternsCompiled []*regexp.Regexp `mapstructure:"-"` // compiled from Patterns during validation

	// default port for database communitation if not specified elsewhere
	DefaultPort int32 `mapstructure:"default_port"`
}

type DomainRule struct {
	DatabaseType string `mapstructure:"database_type"`
	// communication endpoint must match at least one of these patterns
	Patterns         []string         `mapstructure:"patterns"`
	PatternsCompiled []*regexp.Regexp `mapstructure:"-"` // compiled from Patterns during validation

	// in case more DomainRules match, this one will be preferred to be found in service name or endpoint self
	DomainHints []string `mapstructure:"domain_hints"`
}

func (c *Config) Validate() error {
	if err := c.APIConfig.Validate(); err != nil {
		return err
	}

	if c.Interval == 0 {
		c.Interval = defaultInterval
	}

	// validate that rules doesn't have databaseType empty
	if c.Database != nil {
		if err := ValidateDatabaseDiscovery(c.Database); err != nil {
			return err
		}
	}

	return nil
}

func ValidateDatabaseDiscovery(databaseDiscovery *DatabaseDiscoveryConfig) error {
	for i := range databaseDiscovery.ImageRules {
		r := &databaseDiscovery.ImageRules[i]
		if r.DatabaseType == "" {
			return errors.New("database_type must be specified for all image_rules")
		}

		if len(r.Patterns) == 0 {
			return errors.New("at least one match pattern must be specified for all image_rules")
		}

		r.PatternsCompiled = make([]*regexp.Regexp, len(r.Patterns))
		for j, pattern := range r.Patterns {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			r.PatternsCompiled[j] = compiled
		}
	}

	for i := range databaseDiscovery.DomainRules {
		r := &databaseDiscovery.DomainRules[i]
		if r.DatabaseType == "" {
			return errors.New("database_type must be specified for all domain_rules")
		}
		if len(r.Patterns) == 0 {
			return errors.New("at least one match pattern must be specified for all domain_rules")
		}

		r.PatternsCompiled = make([]*regexp.Regexp, len(r.Patterns))
		for j, pattern := range r.Patterns {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			r.PatternsCompiled[j] = compiled
		}
	}
	return nil
}

func (c *Config) getClient() (k8s.Interface, error) {
	if c.makeClient != nil {
		return c.makeClient()
	}
	return k8sconfig.MakeClient(c.APIConfig)
}

// listPods lists all pods across all namespaces using the typed client.
func (c *Config) listPods(ctx context.Context, client k8s.Interface) ([]corev1.Pod, error) {
	pl, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pl.Items, nil
}

// listServices lists all services across all namespaces using the typed client.
func (c *Config) listServices(ctx context.Context, client k8s.Interface) ([]corev1.Service, error) {
	sl, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return sl.Items, nil
}
