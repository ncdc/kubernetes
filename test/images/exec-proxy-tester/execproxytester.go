/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
 * NOTE: This is unsafe to run outside of controlled testing infrastructure!
 *
 * Example curl usage:
 *  curl -X POST -F "file=@/path/to/kubectl" http://127.0.0.1:8080/kubectl/upload
 *  curl -X POST -d "apiserver=http://172.17.42.1:8080&proxy=127.0.0.2&namespace=test&podname=nginx" http://127.0.0.1:8080/kubectl/exec
 *  curl -X POST http://127.0.0.1:8080/kubectl/die
 */

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var KubectlFile string

// requireParam attempts to get the parameter and return it. If it is missing a notice
// is sent back through the pipeline and the caller is notified.
func requireParam(w http.ResponseWriter, r *http.Request, param string) (string, bool) {
	paramValue := r.FormValue(param)
	if paramValue != "" {
		return paramValue, true
	}
	w.WriteHeader(http.StatusNotAcceptable)
	w.Write([]byte(param + " must be set"))
	log.Printf("No %s was passed in. Rejected.", param)
	return "", false
}

// handleKubectlUpload accepts and writes out a kubectl binary
func handleKubectlUpload(w http.ResponseWriter, r *http.Request) {
	if KubectlFile != "" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Binary already uploaded"))
		log.Printf("An attempt to overwrite the binary was rejected.")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to upload file."))
		log.Printf("Unable to upload file: %s", err)
		return
	}
	defer file.Close()

	f, err := ioutil.TempFile("/", "kubectl")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to open file for write."))
		log.Printf("Unable to open file for write: %s", err)
		return
	}
	defer f.Close()
	if _, err = io.Copy(f, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to write file."))
		log.Printf("Unable to write file: %s", err)
		return
	}
	KubectlFile = f.Name()
	err = os.Chmod(KubectlFile, 0700)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to chmod file."))
		log.Printf("Unable to chmod file: %s", err)
		return
	}
	log.Printf("Wrote kubectl to %s", KubectlFile)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Created"))
}

// handleKubectlExec runs the kubectl test command on the pod
func handleKubectlExec(w http.ResponseWriter, r *http.Request) {
	if KubectlFile == "" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Upload kubectly first"))
		log.Printf("An to run a binary without uploading first was made and rejected.")
		return
	}

	// Required params
	envVar, ok := requireParam(w, r, "envvar")
	if !ok {
		return
	}
	apiServerAddr, ok := requireParam(w, r, "apiserver")
	if !ok {
		return
	}
	proxyAddr, ok := requireParam(w, r, "proxy")
	if !ok {
		return
	}
	namespace, ok := requireParam(w, r, "namespace")
	if !ok {
		return
	}
	podname, ok := requireParam(w, r, "podname")
	if !ok {
		return
	}

	log.Printf("Starting execution tests with proxyAddr=%s, namespace=%s, podname=%s", proxyAddr, namespace, podname)
	w.Header().Add("Content-Type", "text/plain")
	log.Printf("Executing with %s=%s", envVar, proxyAddr)
	// Execute the command
	cmd := exec.Command(KubectlFile, "--server="+apiServerAddr, "--namespace="+namespace, "exec", podname, "echo", "running", "in", "container")
	log.Printf("Running: %#v", cmd)
	// Set the environment variable
	cmd.Env = []string{envVar + "=" + proxyAddr}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Unable to execute kubectl: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Print("Execution tests finished")
		w.WriteHeader(http.StatusCreated)
	}
	w.Write(out.Bytes())
}

// handleKubectlDie kills the server
func handleKubectlDie(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Only POST is implemented."))
		log.Print("A non-POST request was recieved")
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Dying..."))
	log.Fatal("Got the request to die.")
}

// postOnly is middleware which enforces handling POST only
func postOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte("Only POST is implemented."))
			log.Print("A non-POST request was recieved")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// main is the main function called via the command line
func main() {
	// Handlers
	http.Handle("/kubectl/upload", postOnly(http.HandlerFunc(handleKubectlUpload)))
	http.Handle("/kubectl/exec", postOnly(http.HandlerFunc(handleKubectlExec)))
	http.Handle("/kubectl/die", postOnly(http.HandlerFunc(handleKubectlDie)))
	// Run the server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
