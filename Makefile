include Makeroutines.mk

VERSION=$(shell git rev-parse HEAD)
DATE=$(shell date +'%Y-%m-%dT%H:%M%:z')
COVER_DIR=/tmp/

# run all tests
define test_only
	@echo "# running unit tests"
	@go test ./logging/logrus
	@go test ./db/keyval/etcdv3
	@go test ./db/keyval/redis
	@go test ./messaging/kafka/client
    @go test ./messaging/kafka/mux
    @go test ./utils/addrs
    @go test ./core
    @echo "# done"
endef

# run all tests with coverage
define test_cover_only
	@echo "# running unit tests with coverage analysis"
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./logging/logrus
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./db/keyval/etcdv3
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit3.out ./messaging/kafka/client
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit4.out ./messaging/kafka/mux
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit5.out ./utils/addrs
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit6.out ./core
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit7.out ./db/keyval/redis
    @echo "# merging coverage results"
    @cd vendor/github.com/wadey/gocovmerge && go install -v
    @gocovmerge ${COVER_DIR}coverage_unit1.out ${COVER_DIR}coverage_unit2.out ${COVER_DIR}coverage_unit3.out ${COVER_DIR}coverage_unit4.out ${COVER_DIR}coverage_unit5.out ${COVER_DIR}coverage_unit6.out ${COVER_DIR}coverage_unit7.out > ${COVER_DIR}coverage.out
    @echo "# coverage data generated into ${COVER_DIR}coverage.out"
    @echo "# done"
endef

# run all tests with coverage and display HTML report
define test_cover_html
    $(call test_cover_only)
    @go tool cover -html=${COVER_DIR}coverage.out -o ${COVER_DIR}coverage.html
    @echo "# coverage report generated into ${COVER_DIR}coverage.html"
    @go tool cover -html=${COVER_DIR}coverage.out
endef

# run all tests with coverage and display XML report
define test_cover_xml
	$(call test_cover_only)
    @gocov convert ${COVER_DIR}coverage.out | gocov-xml > ${COVER_DIR}coverage.xml
    @echo "# coverage report generated into ${COVER_DIR}coverage.xml"
endef

# run code analysis
define lint_only
    @echo "# running code analysis"
    @./scripts/golint.sh
    @./scripts/govet.sh
    @echo "# done"
endef

# build examples only
define build_examples_only
    @echo "# building examples"
    @cd db/keyval/etcdv3/examples && make build
    @cd db/keyval/redis/examples && make build
    @cd logging/logrus/examples && make build
    @cd messaging/kafka/examples && make build
    @echo "# done"
endef

# clean examples only
define clean_examples_only
    @echo "# cleaning examples"
    @cd db/keyval/etcdv3/examples && make clean
    @cd db/keyval/redis/examples && make clean
    @cd logging/logrus/examples && make clean
    @cd messaging/kafka/examples && make clean
    @echo "# done"
endef

# build all binaries
build:
	$(call build_examples_only)

# install dependencies
install-dep:
	$(call install_dependencies)

# update dependencies
update-dep:
	$(call update_dependencies)

# run tests
test:
	$(call test_only)

# run tests with coverage report
test-cover:
	$(call test_cover_only)

# run tests with HTML coverage report
test-cover-html:
	$(call test_cover_html)

# run tests with XML coverage report
test-cover-xml:
	$(call test_cover_xml)

# run & print code analysis
lint:
	$(call lint_only)


# clean
clean:
	@echo "# cleanup completed"
	$(call clean_examples_only)

# run all targets
all:
	$(call lint_only)
	$(call test_only)
	$(call install_only)

.PHONY: build update-dep install-dep test lint clean
