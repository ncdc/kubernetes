// +build !windows

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package term

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/kubernetes/pkg/util/runtime"
)

func monitorResizeEvents(in uintptr, resizeEvents chan<- Size, stop chan struct{}) {
	go func() {
		defer runtime.HandleCrash()

		winch := make(chan os.Signal, 1)
		signal.Notify(winch, syscall.SIGWINCH)
		defer signal.Stop(winch)

		for {
			select {
			case <-winch:
				size := GetSize(in)
				if size == nil {
					return
				}

				// try to send size
				select {
				case resizeEvents <- *size:
					// success
				default:
					// not sent
				}
			case <-stop:
				return
			}
		}
	}()
}
