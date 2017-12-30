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

// +build !node

package nodego

import (
	"flag"
	"log"
	"net"
	"net/http"
)

var address = flag.String("addr", ":8080", "host and port number")

// TakeOver listens and servers http.DefaultServeMux on the address passed by a
// command line flag.
func TakeOver() {
	lis, err := net.Listen("tcp", *address)
	if err != nil {
		panic(err)
	}

	log.Println("listening on", lis.Addr().String())

	if err := http.Serve(lis, nil); err != nil {
		panic(err)
	}
}
