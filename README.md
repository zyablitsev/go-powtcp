# go-powtcp (Proof of Concept)

The service is a tcp-server with protection against DDOS attacks using [Proof of Work](https://en.wikipedia.org/wiki/Proof_of_work) algorithm via challenge-response protocol.

The client sends a request with the `proto.RequestServiceType` message type to the server, in response the server returns a challenge with the `proto.RequestChallengeType` message type to the client. The `proto.RequestChallengeType` message type also includes parameters for the client to perform the required work. The task to be solved is unique and has a lifetime, after which it ceases to be relevant for execution. Having performed the necessary work, the client sends a `ResponseChallengeType` message to the server with a proof of the completed work. If the solution provided by the client is correct, the server returns a quote to the client with a `proto.ResponseServiceType` message.

To prevent exploitation of the server by repeatedly sending the same solution, tasks are one-shot. A single solution can be used only once to retrieve a quote. The complexity of work tasks issued by the server is adaptive and is adjusted depending on the time spent by clients on the solution, the number of active issued tasks and the number of requests within a fixed window.

Key points:

* tcp server protected from ddos attacks with [Proof of Work](https://en.wikipedia.org/wiki/Proof_of_work);
* the challenge-response protocol is used;
* sha-256 hash function is used, as it secure and modern algorithm proven in the bitcoin network;
* after pow-verification server sends one of quotes from "word of wisdom" book.

# Run in docker

1. create config from example
```bash
~$ cp config.docker_example config
~$ vi config
```

2. run build
```bash
~$ make build-amd64
```

3. build docker-image
```bash
~$ make docker-build
```

4. create docker network
```bash
~$ make docker-network
```

5. run server
```bash
~$ make docker-server-run
```

5. run client
```bash
~$ make docker-client-run
```

# Testing

```bash
~$ make test
```

# License

**MIT**
