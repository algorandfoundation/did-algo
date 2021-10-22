# algo-node

The Algorand network is comprised of two distinct types of nodes, relay
nodes, and non-relay nodes. Relay nodes are primarily used for communication
routing to a set of connected non-relay nodes. Relay nodes communicate with
other relay nodes and route blocks to all connected non-relay nodes.
Non-relay nodes only connect to relay nodes and can also participate in
consensus. Non-relay nodes may connect to several relay nodes but never
connect to another non-relay node.

In addition to the two node types, nodes can be configured to be archival and
indexed. Archival nodes store the entire ledger and if the indexer is turned on,
the search range via the API REST endpoint is increased.

This chart facilitate the process of deploying both node types.

More information:
[https://developer.algorand.org/docs/run-a-node](https://developer.algorand.org/docs/run-a-node)
