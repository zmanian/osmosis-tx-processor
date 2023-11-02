
# Osmosis Transaction Processor

This is a demonstration repo of a side car process that a validator should run to ingest and prioritize transactions.

The basic idea is as follows:

The validator runs an HTTP sever with concurrent priority queue.

Wallets, exchanges etc submit their tx to the http server.

The http server responds with what fee they need to pay to get in the next block.

Comet calls out to this process during ReapMaxBytes/ PrepareProposal and gets the bytes the validator commited to proposing in the next block. 




