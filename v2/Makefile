
.PHONY: test
test:
	go test $$(go list ./... | grep -v /NOTICES/)

.PHONY: test-debug
test-debug:
	go test -v $$(go list ./... | grep -v /NOTICES/)

.PHONY: build-linux
build-linux: clean
	mkdir -p dist/linux
	GO111MODULE=on \
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -tags netgo -ldflags '-extldflags "-static"' -o dist/linux/go-licenses github.com/google/go-licenses/v2

.PHONY: build-darwin
build-darwin: clean
	mkdir -p dist/darwin
	GO111MODULE=on \
	CGO_ENABLED=0 \
	GOOS=darwin \
	GOARCH=amd64 \
	go build -tags netgo -ldflags '-extldflags "-static"' -o dist/darwin/go-licenses github.com/google/go-licenses/v2

.PHONY: dist
dist: dist-linux dist-darwin

.PHONY: dist-linux
dist-linux: build-linux
	mkdir -p dist/linux
	cp -r NOTICES dist/linux/
	cp -r third_party/google/licenseclassifier/licenses dist/linux/
	tar -C dist/linux -czf dist/go-licenses-linux.tar.gz \
		go-licenses \
		licenses \
		NOTICES

.PHONY: dist-darwin
dist-darwin: build-darwin
	mkdir -p dist/darwin
	cp -r NOTICES dist/darwin/
	cp -r third_party/google/licenseclassifier/licenses dist/darwin/
	tar -C dist/darwin -czf dist/go-licenses-darwin.tar.gz \
		go-licenses \
		licenses \
		NOTICES

.PHONY: install
install: dist-linux
	cp dist/go-licenses-linux.tar.gz ~/bin/
	cd ~/bin && tar xvf go-licenses-linux.tar.gz && rm -rf NOTICES && rm go-licenses-linux.tar.gz

.PHONY: clean
clean:
	rm -rf dist/linux

.PHONY: upload
upload:
	gsutil cp dist/go-licenses-linux.tar.gz gs://gongyuan-dev/licenses/go-licenses.tar.gz

.PHONY: csv
csv: dist-linux
	dist/linux/go-licenses csv -v 4

.PHONY: save
save:
	dist/linux/go-licenses save -v 4
