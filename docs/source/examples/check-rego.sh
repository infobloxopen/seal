#!/bin/bash -xe

function usage {
  echo -e "$@" >&2
  cat <<EOF >&2
USAGE:
    $(basename $0) {path/to/dir/with/*.rego}

EXAMPLE:
    $(basename $0) petstore
EOF
  exit 2
}

IMAGE="${IMAGE:-openpolicyagent/opa:latest}"

function die { echo -e "$@" >&2; exit 2; }

[[ -z "$1" ]] && usage "rego directory must be specified as parameter 1"
[[ ! -d "$1" ]] && usage "rego directory must be specified as parameter 1"

TOP=$(cd $1 && /bin/pwd)

cd ${TOP} || die "unable to cd to dir: ${TOP}"

docker run -v ${TOP}:/data -w /data "${IMAGE}" fmt -w .

case "$(basename $0)" in
bench-*)
    docker run --rm -v ${TOP}:/data -w /data "${IMAGE}" test -v --bench --count=${OPA_BENCH_COUNT:-2} -t ${OPA_BENCH_TIMEOUT:-10s} *.rego *.json
    ;;
*)
    docker run --rm -v ${TOP}:/data -w /data "${IMAGE}" test -v *.rego *.json
    ;;
esac

git diff --exit-code *.rego || die "opa formatting found whitespace changes, please commit"

