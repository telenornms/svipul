# PREFIX is the prefix on the targetsystem
# DESTDIR can be used to prefix ALL paths, e.g., to do a dummy-install in a
# fake root dir, e.g., for building packages. Users mainly want PREFIX

PREFIX=/usr/local
DOCDIR=${PREFIX}/share/doc/svipul

GIT_DESCRIBE:=$(shell git describe --always --tag --dirty)
VERSION_NO=$(shell echo ${GIT_DESCRIBE} | sed s/[v-]//g)
OS:=$(shell uname -s | tr A-Z a-z)
ARCH:=$(shell uname -m)

all: worker addjob

worker: $(wildcard *.go */*.go */*/*.go go.mod)
	@echo 🤸 go build !
	@go build -ldflags "-X main.versionNo=${VERSION_NO}" -o worker ./cmd/worker

addjob: $(wildcard *.go */*.go */*/*.go go.mod)
	@echo 🤸 go build addjobb !
	@go build -ldflags "-X main.versionNo=${VERSION_NO}" -o addjob ./cmd/addjob

notes: docs/NEWS
	@echo ⛲ Extracting release notes.
	@./build/release-notes.sh $$(echo ${GIT_DESCRIBE} | sed s/-dirty//) > notes

install: worker
	@echo 🙅 Installing
	@install -D -m 0755 worker ${DESTDIR}${PREFIX}/bin/svipul
	@install -D -m 0644 skogul/default.json ${DESTDIR}/etc/svipul/output.d/default.json
	@cd docs; \
	find . -type f -exec install -D -m 0644 {} ${DESTDIR}${DOCDIR}/{} \;
	@install -D -m 0644 README.rst LICENSE -t ${DESTDIR}${DOCDIR}/

build/redhat-svipul.spec: build/redhat-svipul.spec.in FORCE
	@echo  ❕Building spec-file
	@cat $< | sed "s/xxVxx/${GIT_DESCRIBE}/g; s/xxARCHxx/${ARCH}/g; s/xxVERSION_NOxx/${VERSION_NO}/g" > $@
	@if [ ! -f /etc/redhat-release ]; then echo 🆒 Adding debian-workaround for rpm build; sed -i 's/^BuildReq/\#Debian hack, auto-commented out: BuildReq/g' $@; fi

rpm: build/redhat-svipul.spec
	@echo 🎇 Triggering huge-as-heck rpm build
	@mkdir -p rpm-prep/BUILDROOOT
	@DEFAULT_UNIT_DIR=/usr/lib/systemd/system ;\
	RPM_UNIT_DIR=$$(rpm --eval $%{_unitdir}) ;\
	if [ "$${RPM_UNIT_DIR}" = "$%{_unitdir}" ]; then \
	    echo "😭 _unitdir not set, setting _unitdir to $$DEFAULT_UNIT_DIR"; \
	    rpmbuild --quiet --bb \
	        --nodebuginfo \
	    	--build-in-place \
		--define "_rpmdir $$(pwd)" \
		--define "_topdir $$(pwd)" \
		--define "_unitdir $$DEFAULT_UNIT_DIR" \
		--buildroot "$$(pwd)/rpm-prep/BUILDROOT" \
		build/redhat-svipul.spec; \
	else \
	    rpmbuild --quiet --bb \
	        --nodebuginfo \
	    	--build-in-place \
		--define "_rpmdir $$(pwd)" \
		--define "_topdir $$(pwd)" \
		--buildroot "$$(pwd)/rpm-prep/BUILDROOT" \
		build/redhat-svipul.spec; \
	fi
	@cp x86_64/svipul-${VERSION_NO}-1.x86_64.rpm .
	@echo ⭐ RPM built: ./svipul-${VERSION_NO}-1.x86_64.rpm

clean:
	@rm -f worker addjob

check: test fmtcheck vet

mibs:
	@echo ✊ Grabbing mibs
	@tools/get_mibs.sh
vet:
	@echo 🔬 Vetting code
	@go vet ./...

fmtcheck:
	@echo 🦉 Checking format with gofmt -d -s
	@if [ "x$$(find . -name '*.go' -not -wholename './gen/*' -and -not -wholename './vendor/*' -exec gofmt -d -s {} +)" != "x" ]; then find . -name '*.go' -not -wholename './gen/*' -and -not -wholename './vendor/*' -exec gofmt -d -s {} +; exit 1; fi

fmtfix:
	@echo 🎨 Fixing formating
	@find . -name '*.go' -not -wholename './gen/*' -and -not -wholename './vendor/*' -exec gofmt -d -s -w {} +

test:
	@echo 🧐 Testing, without SQL-tests
	@go test -short ./...

bench:
	@echo 🏋 Benchmarking
	@go test -run ^Bench -benchtime 1s -bench Bench ./... | grep Benchmark

covergui:
	@echo 🧠 Testing, with coverage analysis
	@go test -short -coverpkg ./... -covermode=atomic -coverprofile=coverage.out ./...
	@echo 💡 Generating HTML coverage report and opening browser
	@go tool cover -html coverage.out

FORCE:

.PHONY: clean test bench help install rpm release
