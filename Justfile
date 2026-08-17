set default-list := true
set minimum-version := "1.55.0"


##############################################################################
# Source control.
##############################################################################

[group("source")]
[doc("push a branch to both github and codeberg")]
push branch:
	git push github {{branch}}
	git push codeberg {{branch}}


[group("source")]
[doc("tag a release version with message")]
tag version message:
	git tag -a {{version}} -m "{{message}}"


##############################################################################
# Releasing.
##############################################################################

[group("release")]
[doc("tag and release a new version")]
[confirm("Tags are immutable, are you sure?")]
release version message: (tag version message) (push version)
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
