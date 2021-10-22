# algo-indexer

The indexer enables searching the blockchain for transactions, assets,
accounts, and blocks with various criteria. It runs as an independent
process that must connect to a PostgreSQL compatible database that
contains the ledger data. The database is populated by the indexer
which connects to an Algorand node and processes all the ledger data.

The Indexer primarily provides two services, loading a PostgreSQL database
with ledger data and supplying a REST API to search this ledger data. You
can set the Indexer to point at a database that was loaded by another instance
of the Indexer.

More information:
[https://developer.algorand.org/docs/run-a-node/setup/indexer/](https://developer.algorand.org/docs/run-a-node/setup/indexer/)
