# 10. Deployment Slots for Zero Downtime Deploys for the Web App

Date: 2024-09-03

## Decision
1. We will use Azure Web App Deployment Slots to facilitate zero-downtime deploys of the SFTP Ingestion Service web app.
2. Because the ingestion service is queue-driven and in order to keep both the pre-live and production slots healthy,
we will use `sticky_settings` to keep queue configuration on only the production slot in each environment.

## Status

Accepted.

## Context
1. Even though the Ingestion Service's queue-driven workflow is resilient to small downtimes, implementing zero-downtime
deploys is a standard best practice. Using Azure Deployment Slots also lets us have fast and easy rollbacks in addition to
zero-downtime deployment, and is consistent with the workflow we're using in TI.
2. Because the Ingestion Service is queue-driven, turning 'off' the pre-live slot (which routes http traffic) doesn't
stop it from reading queues. To prevent actions from being duplicated, we're keeping queue configuration settings only
on the production/live slot, which will leave the pre-live slot running and healthy, but not active.

Even though there are some significant downsides to Deployment Slots, they're Azure's recommended
approach to zero-downtime deploys (ZDD), and they're lower effort and lower risk than the alternatives.
Other options to achieve ZDD are Kubernetes (significantly more complexity and effort), creating
our own custom deploy system (significantly more complexity, effort, and risk), or switching to
a cloud service provider that makes this easier, like AWS (not currently in scope as an option).

## Impact
### Positive
- **Zero-downtime deploys**: Zero-downtime deploys are a best practice.
- **Easy rollback**: Deployment slots make it easy to roll back to the previous version of the
  app if we find errors after deploy.
- **Consistency**: Deployment Slots are an Azure feature specifically designed to enable
  zero-down time deployment. We use deployment slots in all ingestion service environments and
  in the Trusted Intermediary web app.

### Negative
- **Incomplete support for Linux**: The auto-swap feature is not available for Linux-based web apps like ours.
  so we had to include an explicit swapping step in our updated deployment process.
- **Opaque responses from `az webapp deployment slot swap` CLI**: When there are issues swapping slots, the CLI doesn't
  return any details about the issue. The swapping operation can also take as much as 20 minutes
  to time out if there's a silent failure, which slows down deploy and validation.
- **Steep learning curve**: Most of the official docs and unofficial resources
  (such as blogs and tutorials) for deployment slots are written for people using Windows
  servers and Microsoft-published programing languages. This lack of support for other platforms
  and languages means a lot more trial and error is involved.

### Risks
- Because of the incomplete support for and documentation of our usecase, we may not have
  chosen the optimal implementation of this feature. It may also be time-consuming to
  troubleshoot if we run into future issues.
- Future developers may be confused by which settings should be `sticky` and which should not.
