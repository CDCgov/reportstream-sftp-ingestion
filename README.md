# ReportStream SFTP

## Requirements
Go installed on your machine

## Using and Running
To run the application use the below command

```shell
go run .
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


### Testing

#### Unit Tests
```shell
make unitTests
```

#### End-to-end Tests

#### Load Testing


### Deploying

#### Environments

##### Internal

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

## Related documents

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
