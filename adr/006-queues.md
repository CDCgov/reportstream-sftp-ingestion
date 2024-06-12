# 6. Use events and queues for blob trigger

Date: 2024-06-04

## Decision

When a file arrives in our blob storage, we will use Azure EventGrid to add a file created event to a storage queue.
The SFTP ingestion service will read the event from the queue and handle it (read the file and send it to ReportStream).

## Status

Accepted.

## Context

We needed a way to trigger calling report stream when a file arrives in our blob storage container. Two common ways to
handle this are API endpoints and queues. We selected queues because they include automatic retries and dead lettering
(which increases visibility into what has happened with a message, and makes it easier to replay an event).

The queue message contents are an Azure `Blob Created Event`. The message contains the file name and other metadata,
but does _not_ include any of the file contents and should not contain PHI (unless it is part of the filename).

## Resources
- [Intro to Azure Storage Queues](https://learn.microsoft.com/en-us/azure/storage/queues/storage-queues-introduction)
- [What is Azure Event Grid?](https://learn.microsoft.com/en-us/azure/event-grid/overview)
- [Blob Created Event schema](https://learn.microsoft.com/en-us/azure/event-grid/event-schema-blob-storage?toc=%2Fazure%2Fstorage%2Fblobs%2Ftoc.json&tabs=cloud-event-schema#microsoftstorageblobcreated-event)
