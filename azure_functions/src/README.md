#ReportStream SFTP Functions

## Requirements
[Node](https://nodejs.org/en/download/package-manager/current) version >= 20

[Typescript](https://www.typescriptlang.org/download/) version >= 4.0.0
Azure Functions Core Tools - run `npm install -g azure-functions-core-tools`

## Using and Running

You can run the Azure Function either in Docker using `docker-compose` or from the commands
in `package.json`. Both ways of running the function use the same port, so you can only run
it one way or the other.

### To run outside of docker:

To install dependencies and run the application use the below command

```shell
cd ./azure_functions
npm install
npm run start
```

Other commands can be found in `package.json`


### Verifying Deploys
`functions-deploy.yml` is the CI/CD file responsible for deploying the function into Azure.

In Azure Portal, find your function and look at Logs->Traces to see deploy details.
Typing `trace` into the query console will return the logs for the function you have selected.

As of 7/15/24, the GitHub Action step to deploy the function will silently fail in some
circumstances. If your function depends on e.g. an environment variable that's not set
or is set to an invalid value, the Action will appear to complete successfully, but your
Function App will have not function code in it. It's a good idea to confirm (using the
Azure Portal and/or the `trace` information above) that the function deployed as expected
regardless of the GitHub Action status.
