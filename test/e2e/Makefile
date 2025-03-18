default:
	@echo "hello world"

check-env-gh-token:
	@if [ -z "$${GITHUB_TOKEN}" ] && ! grep -q '^GITHUB_TOKEN=' docker/.env 2>/dev/null; then \
		echo "Error: GITHUB_TOKEN is not set in environment or docker/.env file"; \
		exit 1; \
	else \
		echo "GITHUB_TOKEN is set."; \
	fi

docker-build-all: check-env-gh-token
	docker compose -f docker-compose.build.yml build

docker-build-contracts: check-env-gh-token
	docker compose -f docker-compose.build.yml build hardhat

docker-build-emulator: check-env-gh-token
	docker compose -f docker-compose.build.yml build emulator

docker-all-up:
	docker compose down -v && docker compose up -d

docker-all-down:
	docker compose down -v

docker-all-status:
	docker compose ps -a

docker-eth-down:
	docker compose down eth -v

docker-eth-up:
	docker compose up eth -d

docker-eth-logs:
	docker compose logs eth -f

docker-hardhat-down:
	docker compose down hardhat -v

docker-hardhat-up:
	docker compose up hardhat -d

docker-hardhat-logs:
	docker compose logs -f hardhat

docker-hardhat-shell:
	docker compose exec -it hardhat bash

docker-hardhat-rerun:
	make docker-hardhat-down
	make docker-hardhat-up
	make docker-hardhat-logs

docker-emulator-down:
	docker compose down emulator -v

docker-emulator-up:
	docker compose up emulator -d

docker-emulator-logs:
	docker compose logs -f emulator

docker-emulator-shell:
	docker compose exec -it emulator bash

docker-emulator-rerun:
	make docker-emulator-down
	make docker-emulator-up
	make docker-emulator-logs

docker-test-up:
	docker compose run --rm test

docker-test-up-without-rm:
	docker compose run test

docker-test-shell:
	docker compose run --rm --entrypoint /bin/bash  test

HH_CONTRACTS_DIR = /app/pell-middleware-contracts/deployments/localhost
docker-contracts-save-address:
	@mkdir -p ./_pd-contracts-address/
	docker compose cp hardhat:${HH_CONTRACTS_DIR}/contracts-address-pell.json ../_pd-contracts-address/
	docker compose cp hardhat:${HH_CONTRACTS_DIR}/contracts-address-dvs.json ../_pd-contracts-address/
