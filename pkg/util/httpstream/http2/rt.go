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

package http2

import (
	"net/http"

	"k8s.io/kubernetes/pkg/util/httpstream"
)

type Http2UpgradeRoundTripper struct {
	http.RoundTripper
}

/*
func (rt *Http2RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO what's the best way to clone the request?
	r := *req
	req = &r
	//req.Header.Add(httpstream.HeaderConnection, httpstream.HeaderUpgrade)
	//req.Header.Add(httpstream.HeaderUpgrade, "H2")

	return rt.rt.RoundTrip(req)
}
*/

func (s *Http2UpgradeRoundTripper) NewConnection(resp *http.Response) (httpstream.Connection, error) {
	return &Connection{}, nil
}
