# Secrets

Rotating secrets is tricky and tedious, the purpose of this document is to align contributors with the current state of secrets and ensure they continue to align as secrets are created/renamed or deleted.

## Current Secrets

Below are the secrets that currently exist in Azure KeyVault and what they represent:

- ZIP password: `ca-phl-zip-password-env`
- SFTP starting directory: `ca-phl-sftp-starting-directory-env`
- SFTP server address: `ca-phl-sftp-server-address-env`
- SFTP username: `ca-phl-sftp-user-env`
- SFTP server host public key: `ca-phl-sftp-host-public-key-env`
- SFTP user private key: `ca-phl-sftp-user-credential-private-key-env`
- RS JWT signing key: `ca-phl-reportstream-private-key-env`

### Types

Currently, there are two types of secrets to access our partners. Each secret associated with said service will contain one its name after the partner's name:

- To access ReportStream: `reportstream`
- To access SFTP: `sftp`

There are also two types of keys using RSA:

- Private key, distinguished by .pem ending
- Public key, distinguished by .pem.pub ending

## Naming Convention

The current naming convention for secrets is:

- [partner-name]-[associated-service]-[purpose]

### Past Naming

Previously, the secrets existed in a different name, for prosperity here are the mappings from old to new:

- `ca-dph-zip-password-env` => `ca-phl-zip-password-env`
- `sftp-starting-directory-env` => `ca-phl-sftp-starting-directory-env`
- `sftp-server-address-env` => `ca-phl-sftp-server-address-env`
- `sftp-user-env` => `ca-phl-sftp-user-env`
- `sftp-server-public-key-env` => `ca-phl-sftp-host-public-key-env`
- `sftp-key-env` => `ca-phl-sftp-user-credential-private-key-env`
- `mock-public-health-lab-private-key-env` => `ca-phl-reportstream-private-key-env`

