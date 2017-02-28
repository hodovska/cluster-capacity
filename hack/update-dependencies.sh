#!/bin/bash

rm -rf vendor/

# clear glide cache
glide cc
glide up

stagingpackages=( k8s.io/client-go k8s.io/apiserver k8s.io/apimachinery )

for pack in ${stagingpackages[@]}; do
	rm -rf vendor/$pack/
	cp -R vendor/k8s.io/kubernetes/staging/src/$pack/ vendor/$pack/
	rm -rf vendor/k8s.io/kubernetes/vendor/$pack/
done

# remove conflicting packages
rm -rf vendor/k8s.io/client-go/vendor/k8s.io/apimachinery/
rm -rf vendor/k8s.io/kubernetes/vendor/github.com/golang/glog/
rm -rf vendor/k8s.io/client-go/_vendor/github.com/golang/glog/
