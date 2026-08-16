set default-list := true
set minimum-version := "1.55.0"


##############################################################################
# Releasing.
##############################################################################

[group("release")]
[doc("release the latest version to GitHub")]
[confirm("Have you tagged and pushed to GitHub?")]
release:
	goreleaser release --clean


[group("release")]
[doc("create a local-only test release")]
release-test:
	goreleaser release --snapshot --clean


##############################################################################
# Testing.
##############################################################################

[group("test")]
[doc("run full golanci-lint test suite")]
lint:
	golangci-lint run


[group("test")]
[doc("run all non-integration tests")]
test:
	go test -v -trimpath ./...


[group("test")]
[doc("run all tests")]
test-all:
	go test -v -trimpath -tags integration ./...
