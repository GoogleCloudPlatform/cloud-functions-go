// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package nodego provides utilities for pretending to be node.
package nodego

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
)

func startServer(l net.Listener, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		log.Println(http.Serve(l, nil))
		l.Close()
		wg.Done()
	}()
}

const HTTPTrigger = "/req"

// TakeOver attempts to take over all of node's sockets that were open when it
// execve'd this binary. This binary must have been started by the execer node
// module for this to work.
func TakeOver() {
	fds := os.Args[1:]
	if len(fds) == 0 {
		log.Fatal("No FDs provided.")
	}

	ready := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User function is ready")
	}
	http.HandleFunc("/start", ready)
	http.HandleFunc("/check", ready)
	http.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	var wg sync.WaitGroup
	for _, arg := range fds {
		fd, err := strconv.Atoi(arg)
		if err != nil {
			log.Printf("Error converting arg %q to int: %v", arg, err)
			continue
		}
		f := os.NewFile(uintptr(fd), "")
		l, err := net.FileListener(f)
		f.Close()
		if err != nil {
			log.Println("Error creating FileListener:", err)
			continue
		}
		startServer(l, &wg)
	}

	wg.Wait()
}
