package dsl

import (
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

type CronSourceBuilder struct {
	schedule string
}

func (b CronSourceBuilder) dump() dfv1.Source {
	return dfv1.Source{
		Cron: &dfv1.Cron{
			Schedule: b.schedule,
		},
	}
}

func (b CronSourceBuilder) Cat(name string) StepBuilder {
	return StepBuilder{
		name:    name,
		cat:     &dfv1.Cat{},
		sources: []SourceBuilder{b},
	}
}

func Cron(schedule string) CronSourceBuilder {
	return CronSourceBuilder{schedule: schedule}
}
