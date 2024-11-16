.PHONY: dist test run apitest build-linux
.SILENT: apitest release-rc

test:
	go test ./...

dist: test
	mkdir -p dist
	go build -o dist/auth ./cmd/auth

build-linux: dist
	mkdir -p dist/linux-amd64
	mkdir -p dist/linux-arm64
	docker buildx build --load --platform linux/amd64 -t andrebq/auth:build-linux-amd64 -f ./dockerfiles/Build.Dockerfile .
	docker buildx build --load --platform linux/arm64 -t andrebq/auth:build-linux-arm64 -f ./dockerfiles/Build.Dockerfile .
	docker run --platform linux/amd64 --rm --entrypoint bash andrebq/auth:build-linux-amd64 -c 'cat /usr/local/bin/auth' > ./dist/linux-amd64/auth-linux-amd64
	docker run --platform linux/arm64 --rm --entrypoint bash andrebq/auth:build-linux-arm64 -c 'cat /usr/local/bin/auth' > ./dist/linux-arm64/auth-linux-arm64

release-rc: build-linux
	[[ -n "$(rc)" ]] || { echo "Missing rc=<value> argument"; exit 1; }
	[[ -n "$(semver)" ]] || { echo "Missing rc=<value> argument"; exit 1; }
	rm -rf dist/v$(semver)-rc$(rc)
	mkdir -p dist/v$(semver)-rc$(rc)/
	cp -v dist/linux-amd64/auth-linux-amd64 dist/v$(semver)-rc$(rc)/
	cp -v dist/linux-arm64/auth-linux-arm64 dist/v$(semver)-rc$(rc)/
	gh release create v$(semver)-rc$(rc) dist/v$(semver)-rc$(rc)/*


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
