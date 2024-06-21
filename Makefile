BUILDCMD=CGO_ENABLED=0 go build -mod=vendor
TESTCMD=go test -v -cover -race -mod=vendor
DOCKERNETWORK=pow

# load config parameters
ifneq (,$(wildcard ./config))
	include config
endif

v_server_log_level=$(or ${SERVER_LOG_LEVEL},${server_log_level})
v_server_secret=$(or ${SERVER_SECRET},${server_secret})
v_server_ip=$(or ${SERVER_IP},${server_ip})
v_server_port=$(or ${SERVER_PORT},${server_port})
v_server_rps_target=$(or ${SERVER_RPS_TARGET},${server_rps_target})
v_server_challenge_len=$(or ${SERVER_CHALLENGE_LEN},${server_challenge_len})
v_server_challenge_ttl=$(or ${SERVER_CHALLENGE_TTL},${server_challenge_ttl})
v_server_challenge_pool_cleanup_interval=$(or ${SERVER_CHALLENGE_POOL_CLEANUP_INTERVAL},${server_challenge_pool_cleanup_interval})
v_server_connread_ttl=$(or ${SERVER_CONNREAD_TTL},${server_connread_ttl})
v_server_connwrite_ttl=$(or ${SERVER_CONNWRITE_TTL},${server_connwrite_ttl})

v_client_server_addr=$(or ${CLIENT_SERVER_ADDR},${client_server_addr})
v_client_connread_ttl=$(or ${CLIENT_CONNREAD_TTL},${client_connread_ttl})
v_client_connwrite_ttl=$(or ${CLIENT_CONNWRITE_TTL},${client_connwrite_ttl})

v_server_env=SERVER_LOG_LEVEL="${v_server_log_level}" \
	SERVER_SECRET="${v_server_secret}" \
	SERVER_IP="${v_server_ip}" \
	SERVER_PORT="${v_server_port}" \
	SERVER_RPS_TARGET="${v_server_rps_target}" \
	SERVER_CHALLENGE_LEN="${v_server_challenge_len}" \
	SERVER_CHALLENGE_TTL="${v_server_challenge_ttl}" \
	SERVER_CHALLENGE_POOL_CLEANUP_INTERVAL="${v_server_challenge_pool_cleanup_interval}" \
	SERVER_CONNREAD_TTL="${v_server_connread_ttl}" \
	SERVER_CONNWRITE_TTL="${v_server_connwrite_ttl}"

v_client_env=CLIENT_SERVER_ADDR="${v_client_server_addr}" \
	CLIENT_CONNREAD_TTL="${v_client_connread_ttl}" \
	CLIENT_CONNWRITE_TTL="${v_client_connwrite_ttl}"

test:
	go clean -testcache && \
		${TESTCMD} ./... -run Test

build-server-dev:
	${BUILDCMD} -o ./bin/server_dev cmd/server/main.go

build-client-dev:
	${BUILDCMD} -o ./bin/client_dev cmd/client/main.go

build-dev: \
	build-server-dev \
	build-client-dev

run-server-dev:
	${v_server_env} ./bin/server_dev

run-client-dev:
	${v_client_env} ./bin/client_dev

build-server-amd64:
	GOOS=linux GOARCH=amd64 \
		${BUILDCMD} -o ./bin/server_amd64 cmd/server/main.go

build-client-amd64:
	GOOS=linux GOARCH=amd64 \
		${BUILDCMD} -o ./bin/client_amd64 cmd/client/main.go

build-amd64: \
	build-server-amd64 \
	build-client-amd64

run-server-amd64:
	${v_server_env} ./bin/server_amd64

run-client-amd64:
	${v_client_env} ./bin/client_amd64

docker-network:
	docker network create ${DOCKERNETWORK} 2>/dev/null || true

docker-network-rm:
	docker network rm ${DOCKERNETWORK} 2>/dev/null || true

docker-build:
	docker build -t zyablitsev/go-powtcp .

docker-server-rm:
	docker rm -f powtcp-server 2>/dev/null || true

docker-server-run: docker-server-rm
	docker run \
		-p "0.0.0.0:${v_server_port}:${v_server_port}" \
		-e "SERVER_LOG_LEVEL=${v_server_log_level}" \
		-e "SERVER_SECRET=${v_server_secret}" \
		-e "SERVER_IP=${v_server_ip}" \
		-e "SERVER_PORT=${v_server_port}" \
		-e "SERVER_RPS_TARGET=${v_server_rps_target}" \
		-e "SERVER_CHALLENGE_LEN=${v_server_challenge_len}" \
		-e "SERVER_CHALLENGE_TTL=${v_server_challenge_ttl}" \
		-e "SERVER_CHALLENGE_POOL_CLEANUP_INTERVAL=${v_server_challenge_pool_cleanup_interval}" \
		-e "SERVER_CONNREAD_TTL=${v_server_connread_ttl}" \
		-e "SERVER_CONNWRITE_TTL=${v_server_connwrite_ttl}" \
		--network ${DOCKERNETWORK} \
		--name powtcp-server \
		-d zyablitsev/go-powtcp \
		/server

docker-client-rm:
	docker rm -f powtcp-client 2>/dev/null || true

docker-client-run: docker-client-rm
	docker run \
		-e "CLIENT_SERVER_ADDR=${v_client_server_addr}" \
		-e "CLIENT_CONNREAD_TTL=${v_client_connread_ttl}" \
		-e "CLIENT_CONNWRITE_TTL=${v_client_connwrite_ttl}" \
		--network ${DOCKERNETWORK} \
		--rm zyablitsev/go-powtcp \
		/client
