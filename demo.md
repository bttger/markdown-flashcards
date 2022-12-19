# 1 Databases

## What does ACID stand for?

- Atomicity: transaction's operations are either all executed or none
- Consistency: database always respects all constraints
- Isolation: concurrent transactions do not interfere with each other
- Durability: once a transaction has been committed, it is permanent

## What is read-after-write consistency?

It's the ability (for every client) to view changes to a database immediately after they have been committed.
