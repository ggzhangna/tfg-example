PROJECT_ROOT?=$(shell pwd)
PROJECT_PKG?=tfg-example
TARGET=tfg
TARGET_PKG=$(PROJECT_PKG)
IMAGE_PREFIX=tfg
TARGET_IMAGE=$(IMAGE_PREFIX)/$(TARGET):0.0.1
TARGET_IMAGE_PRD=$(IMAGE_PREFIX)/$(TARGET)-prd:0.0.1

all:image

binary:
	CGO_ENABLED=0 go build -mod vendor  -ldflags "-X main.version=$(GITVERSION)" \
    	 -o dist/$(TARGET) main.go

target:
	mkdir -p $(PROJECT_ROOT)/dist
	docker run --rm -i -v $(PROJECT_ROOT):/$(PROJECT_PKG) \
	  -w /$(PROJECT_PKG) golang:1.12.5 \
      make binary  GITVERSION=`git describe --tags --always --dirty`

image:target
	temp=`mktemp -d` && \
	cp $(PROJECT_ROOT)/dist/$(TARGET) $$temp && cp Dockerfile $$temp && \
	docker build -t $(TARGET_IMAGE) $$temp && \
	rm -rf $$temp && \
	make clean

push-tst:
	docker push $(TARGET_IMAGE)

push-prd:
		docker tag $(TARGET_IMAGE) $(TARGET_IMAGE_PRD) && \
        docker push  $(TARGET_IMAGE_PRD)

dev:image clean
	docker run -it --rm   -p 6000:6000 $(TARGET_IMAGE)  $(TARGET)
clean:
	rm -rf dist


.PHONY: image target clean push binary

