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

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# Update CHANGELOG.md
CHANGELOG_FILE="./CHANGELOG.md"
if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "CHANGELOG.md not found!"
    exit 1
fi
if ! grep -q "## v$VERSION" "$CHANGELOG_FILE"; then
    perl -pi -e "s/^## vNext/## vNext\n\n## v$VERSION/" "$CHANGELOG_FILE"
    echo "CHANGELOG.md updated with version v$VERSION"
else
    echo "CHANGELOG.md already contains 'v$VERSION', no update made."
fi

# Update go.mod files
ALL_GO_MOD=$(find . -name "go.mod" -type f | sort)
for f in $ALL_GO_MOD; do
    perl -pi -e "s|^(\s+github.com/solarwinds/solarwinds-otel-collector-contrib/[^ ]*) v[0-9]+\.[0-9]+\.[0-9]+(\s+// indirect)?$|\1 v$VERSION\2|" "$f"
    echo "References to 'github.com/solarwinds/solarwinds-otel-collector-contrib' in $f updated with version v$VERSION"
done

# update pkg\version\version.go to set the actual release version
GO_VERSION_FILE="./pkg/version/version.go"
if [ ! -f "$GO_VERSION_FILE" ]; then
    echo "version.go not found!"
    exit 1
fi
perl -pi -e "s|^(const Version =) \"[0-9]+\.[0-9]+\.[0-9]+\"$|\1 \"$VERSION\"|" "$GO_VERSION_FILE"
echo "Version.go updated with version $VERSION"
