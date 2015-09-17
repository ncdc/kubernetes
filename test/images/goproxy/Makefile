# TODO: update before merge to gcr.io once accepted for merge and image is uploaded
all: push

TAG = 0.1

goproxy: goproxy.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' ./goproxy.go

image: goproxy
	sudo docker build -t stevemilnerrh/goproxy:$(TAG) .

push: image
	gcloud docker push stevemilnerrh/goproxy:$(TAG)

clean:
	rm -f goproxy
