#!/usr/bin/env bash
# Compilation requires generated admin resources; report that before loading Go modules.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
case "${1:-compile}" in
    modules)
        cd "$root/api-gw"
        go list -mod=readonly -m all > /dev/null
        echo 'Gateway module graph is consistent.'
        exit 0
        ;;
    compile) ;;
    *) echo "Usage: $0 [modules|compile]" >&2; exit 2 ;;
esac
for package in \
    api/build/nexus-client \
    api/build/apis/authentication.admin.nexus.com/v1 \
    api/build/apis/domain.admin.nexus.com/v1 \
    api/build/apis/tenantconfig.admin.nexus.com/v1 \
    api/build/apis/tenantruntime.admin.nexus.com/v1; do
    if [[ ! -d "$root/$package" ]]; then
        printf 'Gateway prerequisite missing: %s\n' "$package" >&2
        printf '%s\n' \
            'Generate the admin API first using api/Makefile (make -C api datamodel_build).' \
            'This needs a compatible compiler image; see api/README.md and docs/testing.md.' >&2
        exit 2
    fi
done
cd "$root/api-gw"
link_flags=${FRAMEWORK_TEST_LDFLAGS:-}
if [[ $(uname -s) == Darwin && -z "$link_flags" ]]; then link_flags=-linkmode=external; fi
go test -mod=readonly -run '^$' -ldflags="$link_flags" ./cmd/api-gw
