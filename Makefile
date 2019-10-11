COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out
COVERAGEREPORT=$(COVERDIR)/report.html

EXAMPLES=$(shell ls ./example/)

$(EXAMPLES): %:
	$(eval EXAMPLE=$*)
	@:

run:
	@test -z "$(EXAMPLE)" && echo "Usage: make [$(EXAMPLES)] run" || go run ./example/$(EXAMPLE)

test:
ifdef CI
	@ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress
else
	@ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress --failFast
endif

test-watch:
	@ginkgo watch --debug -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@ginkgo -r -covermode=count --cover --race --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find . -type f -name *.coverprofile -exec grep -h -v "^mode:" {} >> "${COVERAGEFILE}" \; -exec rm -f {} \;

coverage: coverage-ci
	@sed -i -e "s|_$(CURDIR)/|./|g" "${COVERAGEFILE}"

coverage-html:
	@go tool cover -html="${COVERAGEFILE}" -o $(COVERAGEREPORT)
	@xdg-open $(COVERAGEREPORT) 2> /dev/null > /dev/null

vet:
	@go vet ./...

fmt:
	@go fmt ./...


.PHONY: $(EXAMPLES) run test test-watch coverage coverage-ci coverage-html vet fmt
