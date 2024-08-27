# 7. Code Structure and Organization

Date: 2024-06-06

## Decision

We organize our code by conceptual domain (e.g. `orchestration`, `storage`, and `secrets`) rather than by technology
or deployment location (e.g. `azure` or `local`)

## Status

Accepted.




## Context

We previously had all code that interacted with Azure grouped together in the same package,
and all code that interacted with the local file system in another package.
Once we added queues and queue readers, which trigger the core business logic of the application,
we ended up with circular dependencies between the triggers, the business layer, and the data layer.
By organizing code by purpose instead of technology, we were able to remove the circular dependencies.
