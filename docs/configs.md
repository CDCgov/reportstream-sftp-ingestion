# FAQ
- We use the partner's organization name in ReportStream as the partner ID
- Config files are the partner ID plus `.json`
- Config keys in code are the partner ID
- We don't load configs in the PR environment
- See [The partner settings struct](/src/config/config.go) for the config structure
- Configs load prior to the application running.  Any changes to the config will require a restart of the Azure container to load those changes
- For local non-partner specific testing, we have a Flexion based config that can be used in non-prod environments
- Config files should only contain non-secret values. Secrets will remain in Azure Key Vault
    - secrets will use a consistent naming pattern based on the same partner ID used in config
      (so we can dynamically assemble the key names in code) [see here](../SECRETS.md)
