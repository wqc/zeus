echo "auto makefile ..."
cd $ZEUS_ROOT
cat << END > Makefile

default: all

.PHONY: all
all: eye brain 

.PHONY: pre
pre:
	@./build/run.sh pre

.PHONY: vet
vet: pre
	@./build/run.sh vet

.PHONY: unit_test
unit_test: pre
	@./build/run.sh unit_tests

.PHONY: eye
eye: pre
	@./build/run.sh gobuild eye eye/main.go

.PHONY: brain
brain: pre
	@./build/run.sh gobuild brain brain/main.go

.PHONY: clean
clean:
	@./build/run.sh clean

.PHONY: install
install:
	@./build/run.sh install

.PHONY: gopkg
gopkg: pre
	@./build/run.sh gopkg
END
