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


## Notes
- config files should only contain non-secret values
- secrets will use a consistent naming pattern based on the partner ID used in config (so we can dynamically assemble the key names in code)
- config keys are their file names (minus .json) and match org names in ReportStream

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

- #1082
