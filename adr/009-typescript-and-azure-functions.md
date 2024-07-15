# 9. Typescript and Azure Functions

Date: 2024-07-15

## Decision

We're using Azure Functions for scheduled tasks. Azure Functions support a CRON
expression trigger, which most other Azure resources do not. Azure Webjobs are
supposed to work with CRON triggers, but don't seem to work well with Linux servers since they are in preview.

We're using Typescript for Azure Functions.
Go is not well-supported (it requires special handlers). Of the well-supported
language options, Typescript is lighter-weight than Java or C# and more
familiar to the team than Python.

## Status

Accepted.
