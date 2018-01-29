# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GOBIN = main
OUT = function.zip

all: gcfgo gcfjs
	zip -FS -r $(OUT) $(GOBIN) node_modules index.js package.json -x *build*

gcfgo: FORCE
	GOARCH="amd64" GOOS="linux" go build -tags netgo -tags node $(GOBIN).go

gcfjs: FORCE
	npm install --ignore-scripts --save local_modules/execer

localgo: FORCE
	go build -tags node $(GOBIN).go

localjs: FORCE
	npm install --save local_modules/execer

clean: FORCE
	rm -rf $(GOBIN) $(OUT) node_modules

godev: FORCE
	go run $(GOBIN).go

test: localjs localgo
	node testserver.js

FORCE:
