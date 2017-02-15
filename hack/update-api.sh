#!/usr/bin/env bash

CC_DIR=$GOPATH/src/github.com/kubernetes-incubator/cluster-capacity
CC_PKG=github.com/kubernetes-incubator/cluster-capacity
# create executables for generators
mkdir $CC_DIR/bin

go build -o $CC_DIR/bin/conversion-gen $CC_PKG/vendor/k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
go build -o $CC_DIR/bin/deepcopy-gen $CC_PKG/vendor/k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen


${CC_DIR}/bin/conversion-gen --v 1 --logtostderr \
            --input-dirs github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview/v1alpha1 \
            --output-file-base=zz_generated.conversion

${CC_DIR}/bin/deepcopy-gen --v 1 --logtostderr \
            --input-dirs github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview \
            --input-dirs github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview/v1alpha1 \
            --bounding-dirs github.com/kubernetes-incubator/cluster-capacity \
            --output-file-base=zz_generated.deepcopy

rm -rf ${CC_DIR}/bin
