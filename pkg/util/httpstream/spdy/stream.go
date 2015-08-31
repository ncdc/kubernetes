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

package spdy

import (
	"github.com/docker/spdystream"
	"k8s.io/kubernetes/pkg/util/httpstream"
)

// stream wraps spdystream.Stream to provide better logic for Close().
type stream struct {
	*spdystream.Stream
}

var _ httpstream.Stream = &stream{}

// Close tears down the stream so it can't be used any more. Any pending reads
// are unblocked. This is achieved by issuing a spdystream Reset.
func (s *stream) Close() error {
	return s.Stream.Reset()
}
