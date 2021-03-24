// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package scheduler provides a job scheduler for periodically functions.
package scheduler

import (
	"fmt"
	"github.com/patrickascher/gofer/registry"
	"time"
)

// registry prefix.
const registryPrefix = "gofer:scheduler:"

// Status messages.
const (
	StatusRunning    = "Scheduler is running!"
	StatusNotRunning = "Scheduler is not running"
)

// Pre-defined scheduler providers.
const (
	GoCron = "gocron"
)

type providerFn func(opt interface{}) (Provider, error)

// Provider interface.
type Provider interface {
	// Start the scheduler executor.
	Start()
	// Stop the scheduler executor.
	Stop()
	// Status of the scheduler
	Status() string
	// Jobs of the scheduler will be returned.
	Jobs() []ProviderJobDetail
	// Every will create a new Job with the given interval.
	// int, string and time.duration should be possible.
	Every(interval interface{}) ProviderJob
}

// Job interface.
type ProviderJob interface {
	// Second will be set as unit.
	Second() ProviderJob
	// Minute will be set as unit.
	Minute() ProviderJob
	// Day will be set as unit.
	Day() ProviderJob
	Monday() ProviderJob
	Tuesday() ProviderJob
	Wednesday() ProviderJob
	Thursday() ProviderJob
	Friday() ProviderJob
	Saturday() ProviderJob
	Sunday() ProviderJob
	// At defines the runtime. (Format HH:MM or HH:MM:SS)
	At(string) ProviderJob
	// Week will be set as unit.
	Week() ProviderJob
	// Month will be set as unit.
	// If no day is given, 1 will be set as default.
	Month(dayOfMonth ...int) ProviderJob
	// Name of the job.
	Name(name string) ProviderJob
	// Tag(s) to categorize the job.
	Tag(tag ...string) ProviderJob
	// Singleton will not spawn a new job if the old one is not finished yet.
	Singleton() ProviderJob
	// Do defines the function which should be called. Parameter can be added.
	Do(jobFun interface{}, params ...interface{}) error
}

// JobDetail interface
type ProviderJobDetail interface {
	// Name of the Job
	Name() string
	// Counter of the job runs.
	Counter() int
	// Tags of the job.
	Tags() []string
	// LastRun of the job.
	LastRun() time.Time
	// NextRun of the job.
	NextRun() time.Time
}

// New returns a specific scheduler provider by its name and given options.
// For the specific provider options please check out the provider details.
// If the provider is not registered an error will return.
func New(provider string, options interface{}) (Provider, error) {
	provider = registryPrefix + provider
	// get the registry entry.
	instanceFn, err := registry.Get(provider)
	if err != nil {
		return nil, fmt.Errorf("scheduler: %w", err)
	}

	return instanceFn.(providerFn)(options)
}

// Register a new scheduler provider by name.
func Register(name string, provider providerFn) error {
	return registry.Set(registryPrefix+name, provider)
}
