#!/bin/bash

set -e
ROOT="$( cd "$( dirname $0 )/.." && pwd )"
cd $ROOT
. auto/buildinfo

export GOPATH=$BUILDPATH
export PATH=$PATH:$GOPATH/bin

UNAME=`uname -s`

function pre() {
	if [ $BUILDPATH/src/github.com/zeusship/zeus == $ROOT ]; then
		echo "ERROR: $BUILDPATH equal $ROOT"
	    exit 1
	fi

	mkdir -p $ROOT/bin
	mkdir -p $BUILDPATH/src/github.com/zeusship/
	if [ -d $BUILDPATH/src/github.com/zeusship/zeus ]; then
		rm -r $BUILDPATH/src/github.com/zeusship/zeus
	fi

	if [ -L $BUILDPATH/src/github.com/zeusship/zeus ]; then
		rm -f $BUILDPATH/src/github.com/zeusship/zeus
	fi

	ln -s $ROOT $BUILDPATH/src/github.com/zeusship/zeus
}


function vet() {
	cd $BUILDPATH/src/github.com/zeusship/zeus
	GOPACKGES=`go list ./... | grep -v /vendor/`
	go vet $GOPACKGES 2>&1
}

function unit_tests() {
	cd $BUILDPATH/src/github.com/zeusship/zeus
	test -f coverage.out && rm -f coverage.out >/dev/null 2>&1
	for package in `go list ./... | grep -v 'github.com/zeusship/zeus/vendor'`
	do
		go test -coverprofile=profile.out -covermode=atomic $package
		if [ -f profile.out ]; then
			cat profile.out >> coverage.out
			rm -f profile.out
		fi
	done
}

function gobuild() {
	echo "go build -o $1 $2"
	go build -o $ROOT/bin/$1 $2
}

function gopkg() {
	cd $BUILDPATH/src/github.com/zeusship/zeus
	packages=`go list ./... | grep -v 'zeus/vendor'`
	for package in $packages
	do
		echo "go install $package"
		go install $package
	done

	packages=`cat vendor/vendor.json | grep path | awk -F '"' '{print $4}'`
	for package in $packages
	do
		package="github.com/zeusship/zeus/vendor/$package"
		echo "go install $package"
		go install $package
	done

	gopkg_post
}

function gopkg_post() {
	hostos=$(go env GOHOSTOS)_$(go env GOHOSTARCH)
	for file in `find ${BUILDPATH}/pkg -type file -name '*.a' | grep "/vendor/"`
	do
		file_dir=$(dirname $file)
		new_dir=$BUILDPATH/pkg/$hostos/${file_dir/*\/vendor\//}
		new_file=$BUILDPATH/pkg/$hostos/${file/*\/vendor\//}
		mkdir -p $new_dir
		if [ ! -L $new_file ] && [ ! -f $new_file ] && [ ! -d $new_file ]; then
			ln -s $file $new_file
		fi
	done

	packages=`cat vendor/vendor.json | grep path | awk -F '"' '{print $4}' | sort  -d`
	for package in $packages
	do
		submodule="$BUILDPATH/src/$package"
		parent=`dirname $submodule`
		if [ -L $submodule ] || [ -d $submodule ] || [ -f  $submodule ]; then
			continue
		fi

		if [ -L $parent ]; then
			continue
		fi

		if [ -f $parent ]; then
			continue
		fi

		if [ ! -d $parent ]; then
			mkdir -p $parent
		fi


		ln -s  $ROOT/vendor/$package $submodule
	done
}

function install() {
	test -d ${DESTDIR}${BINPATH} || mkdir -p ${DESTDIR}${BINPATH}
	test -d ${DESTDIR}${CONFPATH} || mkdir -p ${DESTDIR}${CONFPATH}
	cp misc/start-brain.sh ${DESTDIR}${BINPATH}
	cp misc/start-eye.sh ${DESTDIR}${BINPATH}
	cp misc/start.sh ${DESTDIR}${BINPATH}
	cp misc/stop.sh  ${DESTDIR}${BINPATH}
	cp misc/stop-eye.sh ${DESTDIR}${BINPATH}
	cp misc/stop-brain.sh ${DESTDIR}${BINPATH}
}

function clean() {
	rm -f $BUILDPATH/src/github.com/zeusship/zeus
	rm -rf $BUILDPATH/pkg/*
	find $BUILDPATH/src -type link | xargs rm -f
	rm -f Makefile
	rm -rf auto
	rm -rf bin
}

function main () {
	if [[ $# < 1 ]]; then
		targets="pre"
	else
		targets=($@)
	fi

	if [ $1 == "gobuild" ]; then
		$1 $2 $3
	else
		for target in ${targets[@]}; do
			$target
		done
	fi
}

main "$@"
