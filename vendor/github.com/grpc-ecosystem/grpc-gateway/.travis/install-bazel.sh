#!/bin/sh -eu

bazel_version=$1

if test -z "${bazel_version}"; then
	echo "Usage: .travis/install-bazel.sh bazel-version"
	exit 1
fi

if [[ "${TRAVIS_OS_NAME}" == "osx" ]]; then
	OS=darwin
else
	OS=linux
fi

filename=bazel-${bazel_version}-installer-${OS}-x86_64.sh
wget https://github.com/bazelbuild/bazel/releases/download/${bazel_version}/${filename}
chmod +x $filename
./$filename --user
rm -f $filename
