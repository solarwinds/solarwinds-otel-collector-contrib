include Makefile.Common

ALL_SRC := $(shell find . \( -name "*.go" \) \
							-not -path '*generated*' \
							-type f | sort)

.PHONY: ci-check-licenses
ci-check-licenses: add-licenses check-licenses

.PHONY: add-licenses
add-licenses: $(ADDLICENSE)
	@addLicensesOutput=`$(ADDLICENSE) -y "2025" -c "SolarWinds Worldwide, LLC. All rights reserved." -l "apache" \
	    -v $(ALL_SRC) 2>&1`; \
		if [ "$$addLicensesOutput" ]; then \
			echo "Files modified:"; \
			echo "$$addLicensesOutput"; \
			exit 1; \
		else \
			echo "No files modified."; \
		fi

.PHONY: check-licenses
check-licenses:
	@build/check-licenses.sh ${EXPECTED_GO_LICENSE_HEADER} ${EXPECTED_SHELL_LICENSE_HEADER}

.PHONY: prepare-release
prepare-release:
	@build/prepare-release.sh $(version)
