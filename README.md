# CS598FTS Project

## Build

```
make
```

## Run

### Client

**Read**

```sh
./cs598fts client read testkey --id 0 --config config/local.json
# Or
./script/client read testkey --id 0 --config config/local.json
```

**Write**

```sh
./cs598fts client write testkey testval --id 0 --config config/local.json
# Or
./script/client write testkey testval --id 0 --config config/local.json
```

**Benchmark**

```sh
./cs598fts client benchmark --config config/local-3.json --client 1 --request 100000 --workload half-half
```

### Server

```sh
./cs598fts server :8080
# Or
./script/replica :8080
```

### Script for local testing

Start 5 servers locally, 10 client threads, and perform a half-half workload of 5k writes and 5k reads per client thread
```sh
python3 script/local.py 5
```

## Design and Implementation

Our project presents a Multi-Writer Multi-Reader (MWMR) shared register system, which is designed using the ABD algorithm. The system follows a client-server architecture, where the server provides two interfaces, namely `set` and `get`, to the clients. On the server side, each register object is associated with a state that includes a value, a timestamp, and client ID records. When a new `set` request arrives, the server compares its timestamp with the one stored locally and updates its internal state if the incoming request is newer. The server sends an acknowledgment to the client once the update is complete. On receiving a `get` request, the server returns the current state of the register object.

The client supports two basic operations: read and write. The read operation first retrieves the state of the register object from server replicas through the `get` interafce and waits for a response from the majority. It then finds the largest timestamp and corresponding value from the received responses, and invokes the `set` interface on the server with the maximum timestamp + 1 and its own client ID. For write, the operation is similar, except that it uses the new value for the `set` request.

We implemented the system in Golang and utilized the language's built-in rpc library for client-server communication. The communication channel is HTTP-based over a TCP socket, ensuring reliability. Each object's storage has a lock, and access to object state through `set`/`get` interfaces is single-threaded, the server can still process requests to different objects in parallel to improve throughput. On the client side, multiple goroutines are spawned to send requests to all server replicas in parallel, and the client waits for the completion of all requests before deciding whether to proceed or abort with an error, based on whether the completion count is greater than the number of majority (n+1)/2.