#!/bin/bash
make
./cs598fts client benchmark --config config/local.json --client 10 --request 10000 --workload half-half