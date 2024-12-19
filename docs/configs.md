# FAQ

- We don't load configs in the PR environment.  
- See [The partner settings struct](/src/config/config.go) for the config structure
- Configs load prior to the application running.  Any changes to the config will require a restart of the Azure container to load those changes
- For local non-partner specific testing, we have a Flexion based config that can be used in non-prod environments

