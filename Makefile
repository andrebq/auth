.PHONY: dist test run apitest build-linux-amd64
.SILENT: apitest

test:
	go test ./...

dist: test
	mkdir -p dist
	go build -o dist/auth ./cmd/auth

build-linux-amd64:
	mkdir -p dist/linux-amd64
	docker buildx build --load --platform linux/amd64 -t andrebq/auth:build-linux-amd64 -f ./dockerfiles/Build.Dockerfile .
	docker run --platform linux/amd64 --rm --entrypoint bash andrebq/auth:build-linux-amd64 -c 'cat /usr/local/bin/auth' > ./dist/linux-amd64/auth

run: dist
	mkdir -p ./localfiles/var/auth
	cd dist && \
		./auth --data-dir ../localfiles/var/auth/ serve api

testScript?=$(PWD)/internal/e2etests/api_tests.sh
authEndpoint?=http://localhost:18001
apitest: dist
	[[ -n "$(authEndpoint)" ]] || { echo "Missing argument authEndpoint=http:...., please check and try again" 1>&2; exit 1; }
	[[ -n "$(testScript)" ]] || { echo "Missing argument testScript=./path/to/test/script, please check and try again" 1>&2; exit 1; }
	AUTH_ENDPOINT="$(authEndpoint)" bash "$(testScript)"
