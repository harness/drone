// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package services

import (
	"github.com/harness/gitness/internal/services/job"
	"github.com/harness/gitness/internal/services/metric"
	"github.com/harness/gitness/internal/services/pullreq"
	"github.com/harness/gitness/internal/services/trigger"
	"github.com/harness/gitness/internal/services/webhook"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideServices,
)

type Services struct {
	Webhook         *webhook.Service
	PullReq         *pullreq.Service
	Trigger         *trigger.Service
	JobScheduler    *job.Scheduler
	MetricCollector *metric.Collector
}

func ProvideServices(
	webhooksSvc *webhook.Service,
	pullReqSvc *pullreq.Service,
	triggerSvc *trigger.Service,
	jobScheduler *job.Scheduler,
	metricCollector *metric.Collector,
) Services {
	return Services{
		Webhook:         webhooksSvc,
		PullReq:         pullReqSvc,
		Trigger:         triggerSvc,
		JobScheduler:    jobScheduler,
		MetricCollector: metricCollector,
	}
}
