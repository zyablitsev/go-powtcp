# go-powtcp (Proof of Concept)

Implementation of “Word of Wisdom” tcp server.

# Notes

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
