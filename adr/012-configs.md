# 11. Azure Alerts

Date: 2024-12-12

## Decision

We will store partner config settings in an Azure container

## Status

Accepted.

## Context

In order to enable the usage of partner specific settings in the different parts of the app we are going to store the settings
in a config container in our Azure storage.  Each partner will have it's own separate file in the container to minimize potential
blast radius when changing settings.


## Impact

### Positive

- We can continue to meet our partners where they are by having partner specific settings in the app in order to provide any needed customizations.

### Negative

- Some added complexity for the implementation of configs.

### Risks

- None

## Related Issues

- #1082
