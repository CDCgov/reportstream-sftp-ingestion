# Secrets

## Current Secrets

Below are the secrets that currently exist in Azure KeyVault and what they represent:

- ZIP password: `ca-phl-zip-password-env`
- SFTP starting directory: `ca-phl-sftp-starting-directory-env`
- SFTP server address: `ca-phl-sftp-server-address-env`
- SFTP username: `ca-phl-sftp-user-env`
- SFTP server host public key: `ca-phl-sftp-host-public-key-env`
- SFTP user private key: `ca-phl-sftp-user-credential-private-key-env`
- RS JWT signing key: `ca-phl-reportstream-private-key-env`

## Types

Currently, there are these types of keys:

- To access to partners:
  - To access report stream
  - To access SFTP

## Naming Convention

The current naming convention for secrets is:

- partner-name-user-type

## Past Naming

Previously, the secrets existed in a different name, here are the mappings from old to new:

- `ca-dph-zip-password-env` => `ca-phl-zip-password-env`
- `sftp-starting-directory-env` => `ca-phl-sftp-starting-directory-env`
- `sftp-server-address-env` => `ca-phl-sftp-server-address-env`
- `sftp-user-env` => `ca-phl-sftp-user-env`
- `sftp-server-public-key-env` => `ca-phl-sftp-host-public-key-env`
- `sftp-key-env` => `ca-phl-sftp-user-credential-private-key-env`
- `mock-public-health-lab-private-key-env` => `ca-phl-reportstream-private-key-env`

