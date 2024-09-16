# Secrets

Rotating secrets is tricky and tedious, the purpose of this document is to align contributors with the current state of
secrets and ensure they continue to align as secrets are created/renamed or deleted.

## Current Secrets

Below are the secrets that currently exist in Azure KeyVault and what they represent.  The `env` part represents the
environment, such as `dev`, `stg`, etc.

- ZIP password: `ca-phl-zip-password-env`.
- SFTP starting directory: `ca-phl-sftp-starting-directory-env`.
- SFTP server address: `ca-phl-sftp-server-address-env`.
- SFTP username: `ca-phl-sftp-user-env`.
- SFTP server host public key: `ca-phl-sftp-host-public-key-env`.
- SFTP user credential private key: `ca-phl-sftp-user-credential-private-key-env`.
- RS JWT signing key: `ca-phl-reportstream-private-key-env`.

## Naming Convention

The current naming convention for secrets is: [partner-name]-[associated-service]-[purpose]

If we ever have private keys or public keys, include whether it is a private or public key in the [purpose].  See our
existing secrets for inspiration.

## Types

Currently, there are two types of secrets to access our partners. Each secret associated with said service will contain
one its name after the partner's name:

- To access ReportStream: `reportstream` is included in the name.
- To access SFTP: `sftp` is included in the name.

There are also two types of secrets when it comes to connecting to an SFTP server.

### The User Credentials

This is represented as `user-credential-private-key` in our secrets.

This is a private key that the user (us) has and use to authenticate to the SFTP server.  The associated
public key is given to the SFTP server administrator before we try to connect.

### The Server's Host Key

This is represented as `sftp-host-public-key` in our secrets.

This is a public key that the user (us) has and use to ensure we are connecting to the correct SFTP server.  The
associated private key is pre-created by the SFTP server administrator and installed on the SFTP server.  We don't
create the public key, but we can get the public key when we connect to the server or the SFTP server administrator can
give it to us.

## Local Secrets

We put mock secrets into [mock_credentials](./mock_credentials) that are used when running the service locally.  There
are additional secrets in there than are used by our application.  For example, we have a private key for our user
credentials as one of our secrets, but we need an associated public key to be installed on the mock SFTP server for this
to work.  So, we have a public key that isn't used by our service but is used by our mock SFTP server to make the
authentication work.  A similar concept applies to the SFTP host key and could apply to other secrets.
