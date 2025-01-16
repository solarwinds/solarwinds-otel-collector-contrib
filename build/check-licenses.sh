#!/bin/bash
# Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
 
EXPECTED_GO_LICENSE_HEADER=$1
EXPECTED_SHELL_LICENSE_HEADER=$2
ALL_SRC=$(find . \( -name "*.go" -o -name "*.sh" \) \
			     -not -path '*generated*' \
			     -type f | sort)

RC=0
for f in $ALL_SRC; do
    case "$f" in
        *.go)
            if ! diff -q <(head -n 13 "$f") $EXPECTED_GO_LICENSE_HEADER > /dev/null; then
                echo "Missing or incorrect license headers in Go source files!"
                echo "Diff for $f:";
                diff --label="$f" -u <(head -n 13 "$f") $EXPECTED_GO_LICENSE_HEADER;
                RC=1
            fi;
        ;;
        *.sh)
            if ! diff -q <(tail -n +2 "$f" | head -n 13) $EXPECTED_SHELL_LICENSE_HEADER > /dev/null; then
                echo "Missing or incorrect license headers in shell source files!"
                echo "Diff for $f:";
                diff --label="$f" -u <(tail -n +2 "$f" | head -n 13) $EXPECTED_SHELL_LICENSE_HEADER;
                RC=1
            fi;
        ;;
    esac;
done
exit $RC
