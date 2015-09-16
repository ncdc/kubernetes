all: push

TAG = 0.1

execproxytester: execproxytester.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' ./execproxytester.go

image: execproxytester
	sudo docker build -t gcr.io/google_containers/exec-proxy-tester:$(TAG) .

push: image
	gcloud docker push gcr.io/google_containers/exec-proxy-tester:$(TAG)

clean:
	rm -f execproxytester
