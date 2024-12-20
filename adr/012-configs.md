# 12. Partner configuration

Date: 2024-12-12

## Decision

We will store partner config settings in an Azure container

## Status

Accepted.

## Context

In order to enable the usage of partner-specific settings in the different parts of the app, we are going to store the settings
in a config container in our Azure storage account. Each partner will have its own separate file within the container to minimize potential
blast radius when changing settings.

## Impact

### Positive

- We can continue to meet our partners where they are by having partner specific settings in the app in order to provide any needed customizations.
- We can create separate testing config for the Flexion organizations

### Negative

- Some added complexity for the implementation of configs.
- Initial implementation of the config will require either restarting or redeploying the app

### Risks

- None

## Related Issues

- #[1082](https://github.com/CDCgov/trusted-intermediary/issues/1082)
