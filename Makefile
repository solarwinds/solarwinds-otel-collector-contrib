include Makefile.Common

ALL_SRC := $(shell find . \( -name "*.go" \) \
							-not -path '*generated*' \
							-type f | sort)

.PHONY: ci-check-licenses
ci-check-licenses: add-licenses check-licenses

.PHONY: add-licenses
add-licenses: $(ADDLICENSE)
	@addLicensesOutput=`$(ADDLICENSE) -y "2024" -c "SolarWinds Worldwide, LLC. All rights reserved." -l "apache" -v $(ALL_SRC) 2>&1`; \
		if [ "$$addLicensesOutput" ]; then \
			echo "Files modified:"; \
			echo "$$addLicensesOutput"; \
			exit 1; \
		else \
			echo "No files modified."; \
		fi

.PHONY: check-licenses
check-licenses:
	@checkResult=$$(for f in $(ALL_SRC) ; do \
	         if ! diff -q <(head -n 13 $$f) $(EXPECTED_LICENSE_HEADER) > /dev/null; then \
				  echo "Diff for $$f:"; \
				  diff --label $$f -u <(head -n 13 $$f) $(EXPECTED_LICENSE_HEADER); \
		     fi; \
	   done); \
	   if [ -n "$${checkResult}" ]; then \
	           echo "License header check failed:"; \
	           echo "$${checkResult}"; \
	           exit 1; \
	   fi
