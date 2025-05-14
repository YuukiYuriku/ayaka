# simple makefile
NAME=main
COVERAGE_MIN=50.0

## check: Check quality code in sonarqube.
check:
	@sonar-scanner -Dsonar.projectKey="$SONAR_PROJECT_KEY" -Dsonar.sources=. -Dsonar.host.url="$SONAR_HOST" -Dsonar.login="$SONAR_LOGIN"

## test: Run test and enforce go coverage
test:
	@go generate ./...
	@echo "Checking test coverage..."
	@echo "Minimum Coverage: $(COVERAGE_MIN)%"
	@go test ./... -coverprofile cp.out
	@go tool cover -func=cp.out | grep total | awk -v min=$(COVERAGE_MIN) '{ \
		coverage = substr($$3, 1, length($$3)-1) + 0; \
		min += 0; \
		if (coverage < min) { \
			print "Coverage is " coverage "% below required threshold " min "%" ; \
			exit 1; \
		} else { \
			print "Coverage passed threshold: " coverage "%"; \
		} \
	}'

## coverage: Show go coverage
coverage: test
	@echo "coverage details:";
	@go tool cover -func=cp.out

## coverage-web: Show go coverage in web
coverage-web: test
	@go tool cover -html=cp.out -o coverage.html

## bench: Run benchmark test
bench:
	go test -bench=.

## watch: development with air
watch:
	air -c .air.toml

## build: Build binary applications
build:
	@go generate ./...
	@echo building binary to ./bin/${NAME}
	@go build -o ./bin/${NAME} .

## deploy: Deploy binary to server using ansible-playbook
deploy:
	@echo "Command to deploy script distribute atrifacts to cloud, on-prem or kubernetes clusters "

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run with parameter options: "
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
