# ReportStream SFTP

> [!WARNING]
> This application is no longer being developed or maintained.

## Requirements

Go installed on your machine

## A Note on File Encoding
As of October 2024, we are only connecting to CADPH. Their SFTP site uses ISO 8859-1 encoding, which
must be converted to UTF-8 before sending to ReportStream or special characters will be garbled.
Before we can connect with any partners who use another encoding, we'll need to make encoding
conversion configurable.

## Using and Running

To run the application use the below command

```shell
cd ./src/
go run ./cmd/
```

We also have other build/lint/test commands listed in [makefile](Makefile). You can run these in your terminal using the syntax: make (step name)


```shell
make vet
```

## Development

### Additional Requirements

The additional requirements needed to contribute towards development are...

- [Pre-Commit](https://pre-commit.com)
- [Detect-Secrets](https://github.com/Yelp/detect-secrets)
- [Terraform](https://www.terraform.io)
- [Docker](https://www.docker.com/)
- [Azurite](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=visual-studio%2Cblob-storage#install-azurite)

- [Microsoft Azure Storage Explorer - needed for local manual testing](https://azure.microsoft.com/en-us/products/storage/storage-explorer)

### Compiling

```shell
make compile
```

Compiles the binary to the root of the project


### Pre-Commit Hooks

We use [`pre-commit`](https://pre-commit.com) to run [some hooks](./.pre-commit-config.yaml) on every commit.  These
hooks do linting to ensure things are in a good spot before a commit is made.  Please install `pre-commit` and then
install the hooks.

```shell
pre-commit install
```

### Running locally
Run `docker-compose`, which will spin up an Azurite container, an SFTP service, the
Azure Function that triggers polling, and the app. By default, this leaves the ReportStream
URL prefix environment variable empty, and we'll use a mock response rather than calling ReportStream. Uncomment
the `REPORT_STREAM_URL_PREFIX` in [docker-compose.yml](docker-compose.yml) to call locally-running ReportStream instead.


### Testing

#### Unit Tests

```shell
make unitTests
```

To update which files are excluded from the test coverage report update the `codeCoverageFilter.sh`script.

#### Manual local testing
In the cloud, EventGrid monitors the blob storage container and sends file creation events to our queue for the app to read.
In the local Azurite tool, there are no events to connect the blob storage container to the queue.
To mimic the deployed behavior so our app can read queue messages and access the file specified in the message:
1. Upload a file to your local Azurite sftp container
2. In Azure Storage Explorer, find the `message-import-queue` that the service currently reads from
3. Add a file create event message to that queue. You can start with this base message and edit the `subject` to
match your newly-created file
   ```json
   {
   "topic": "/subscriptions/52203171-a2ed-4f6c-b5cf-9b368c43f15b/resourceGroups/csels-rsti-internal-moderate-rg/providers/Microsoft.Storage/storageAccounts/cdcrssftpinternal",
   "subject": "/blobServices/default/containers/sftp/blobs/import/order_message.hl7",
   "eventType": "Microsoft.Storage.BlobCreated",
   "id": "dac45448-001e-0031-7649-b8ad2c06c977",
   "data": {
   "api": "PutBlob",
   "clientRequestId": "d621de4c-d460-4af1-85f9-126d06328012",
   "requestId": "dac45448-001e-0031-7649-b8ad2c000000",
   "eTag": "0x8DC8660D8D1AE8B",
   "contentType": "application/octet-stream",
   "contentLength": 1122,
   "blobType": "BlockBlob",
   "url": "http://127.0.0.1:12000/devstoreaccount1/sftp/import/order_message.hl7",
   "sequencer": "00000000000000000000000000024DA1000000000006ab03",
   "storageDiagnostics": {
   "batchId": "6677b768-3006-0093-0049-b89735000000"
           }
       },
   "dataVersion": "",
   "metadataVersion": "1",
   "eventTime": "2024-06-06T19:42:49.2380563Z"
   }
   ```
4. The app should now read this message and attempt to process it

For the external SFTP call, we've set up a file in docker-compose that's copied to the local SFTP server. The service
then copies it to local Azurite. You can add additional files by placing them in `localdata/data/sftp` before running
`docker-compose`.

As of 7/3/24, when we copy a file from the local SFTP server, we try to unzip it
(using the password in `mock_credentials/mock_ca_dph_zip_password.txt` if it's protected). We then place the unzipped
files into the import folder, and if there are any errors, we upload an error file for the zip. If the original file is
not a zip, we just copy it into the import folder.

#### Manual Cloud Testing

##### Upload to Our Azure Container

To trigger file ingestion in a deployed environment, go to the `cdcrssftp{env}` storage account in the Azure Portal.
In the `sftp` container, upload a file to an `import` folder. If that folder doesn't already exist, you can create
it by going to `Upload`, expanding `Advanced`, and putting `import` in the `Upload to folder` box!
[upload_file.png](docs/upload_file.png)

##### Upload to SFTP Server

Log into CA's SFTP staging environment and drop a file into the `OUTPUT` folder.  You can either wait for one of our
lower environments to trigger, or you can manually trigger the Azure function from the Azure Portal.

The credentials and domain name for CA's SFTP environment can be found in Keybase under CA Info.

To manually trigger the Azure function...

1. If this is in an Azure Entra domain environment, you will need to log in as your -SU account.
2. Go to the `polling-function-{env}` function app in the Azure Portal.
3. Navigate to the CORS section, and add `https://portal.azure.com` as an allowed origin.  Click save.
4. Navigate back to the Overview section, and click on the trigger function.
5. Click on the Test/Run button, and then click on the Run button that pops-up.


#### End-to-end Tests

#### Load Testing


### Deploying

#### Environments

##### Internal

The Internal environment is meant to be the Wild West.  Meaning anyone can push to it to test something, and there is no
requirement that only good builds be pushed to it.  Use the Internal environment if you want to test something in a
deployed environment in a _non-CDC_ Azure Entra domain and subscription.

To deploy to the Internal environment...
1. Check with the team that no one is already using it.
2. [Find the `internal` branch](https://github.com/CDCgov/reportstream-sftp-ingestion/branches/all?query=internal) and delete
   it in GitHub.
3. Delete your local `internal` branch if needed.
   ```shell
   git branch -D internal
   ```
4. From the branch you want to test, create a new `internal` branch.
   ```shell
   git checkout -b internal
   ```
5. Push the branch to GitHub.
   ```shell
   git push --set-upstream origin internal
   ```

Then the [deploy](https://github.com/CDCgov/reportstream-sftp-ingestion/actions/workflows/internal-deploy.yml) will run.
Remember that you now have the `internal` branch checked out locally.  If you make subsequent code changes, you will
make them on the `internal` branch instead of your original branch.

##### Dev

The Dev environment is similar to the Internal environment but deploys to a CDC Azure Entra domain and subscription.  It
is also meant to be the Wild West.  Dev deploys similarly to the Internal environment, but you interact with the
`dev` branch.

##### Staging

The Staging environment is production-like and meant to be stable.  It deploys to a CDC Azure Entra domain and
subscription.  Deployments occur when a commit is made to the `main` branch.  `main` is a protected branch and requires
PR reviews before merge.

##### Prod

The Production environment is the real deal.  It deploys to a CDC Azure Entra domain and subscription.  Deployments
occur when a release is published.

### Secrets

See [SECRETS.md](./SECRETS.md) for a description of our secrets.

## Related documents

* [Azure Functions and Typescript](/azure_functions/src/README.md)
* [Open Practices](/docs/open_practices.md)
* [Rules of Behavior](/docs/rules_of_behavior.md)
* [Thanks and Acknowledgements](/docs/thanks.md)
* [Disclaimer](DISCLAIMER.md)
* [Contribution Notice](CONTRIBUTING.md)
* [Code of Conduct](/docs/code-of-conduct.md)

## CDC Notices

## Public Domain Standard Notice

This repository constitutes a work of the United States Government and is not
subject to domestic copyright protection under 17 USC § 105. This repository is in
the public domain within the United States, and copyright and related rights in
the work worldwide are waived through the [CC0 1.0 Universal public domain dedication](https://creativecommons.org/publicdomain/zero/1.0/).
All contributions to this repository will be released under the CC0 dedication. By
submitting a pull request you are agreeing to comply with this waiver of
copyright interest.

## License Standard Notice

The repository utilizes code licensed under the terms of the Apache Software
License and therefore is licensed under ASL v2 or later.

This source code in this repository is free: you can redistribute it and/or modify it under
the terms of the Apache Software License version 2, or (at your option) any
later version.

This source code in this repository is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the Apache Software License for more details.

You should have received a copy of the Apache Software License along with this
program. If not, see http://www.apache.org/licenses/LICENSE-2.0.html

The source code forked from other open source projects will inherit its license.

## Privacy Standard Notice

This repository contains only non-sensitive, publicly available data and
information. All material and community participation is covered by the
[Disclaimer](DISCLAIMER.md)
and [Code of Conduct](/docs/code-of-conduct.md).
For more information about CDC's privacy policy, please visit [http://www.cdc.gov/other/privacy.html](https://www.cdc.gov/other/privacy.html).

## Records Management Standard Notice

This repository is not a source of government records, but is a copy to increase
collaboration and collaborative potential. All government records will be
published through the [CDC website](http://www.cdc.gov).

## Additional Standard Notices

Please refer to [CDC's Template Repository](https://github.com/CDCgov/template) for more information about [contributing to this repository](https://github.com/CDCgov/template/blob/main/CONTRIBUTING.md), [public domain notices and disclaimers](https://github.com/CDCgov/template/blob/main/DISCLAIMER.md), and [code of conduct](https://github.com/CDCgov/template/blob/main/code-of-conduct.md).
