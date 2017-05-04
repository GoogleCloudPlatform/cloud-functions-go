# Unofficial Native Go Runtime for Google Cloud Functions

_Disclaimer: This is not an official Google product. It is not and will not be maintained by Google, and is not part of Google Cloud Functions project. There is no guarantee of any kind, including that it will work or continue to work, or that it will supported in any way._

## Background

At the time of writing, Google Cloud Functions only supports a Node.js based runtime. Running non-JavaScript code is only officially supported in the form of native Node modules and running subprocesses. Node will always be handling the requests. Wrapping Go code in a native Node module will require writing complicated foreign function interfaces (FFIs). FFIs are notoriously difficult to get right and are an easy way to introduce memory leaks as well as security and stability issues. Using a subprocess is much easier to get right, but it introduces the overhead of interprocess communication (IPC) and context switches. In both cases, the requests are proxied between Node and Go which typically involves extra copies and translations.

## 100% Pure Go Request Path
After an initial bootstrap step, the only thing running is a pure Go binary. This project works by completely replacing the Node process and lets a Go process directly handle requests. There is no cgo and no proxying. This means that it is faster and potentially more secure than other projects which typically to use both.

## How It Works
A native module calls execve(2) without the normal preceding fork(2) call when the user module is imported. Open socket FDs are preserved and handled by the new Go process to ensure a smooth transition. The new Go process then pretends to be Node.

## Requirements
* Linux or macOS development environment (The native Node module depends on Linux/macOS specific behavior.)
* Go 1.5 or above
* Node.js v0.10 or above and node-gyp
* Make and GCC
* If you are using macOS then as of right now you need the FULL xcode installed. Command line tools alone will not work.

## Hello, world!
A demo hello world example is included. To try it out, simply skip to the [Deployment](#deployment) section.

## Making Changes
For normal development, it should only be necessary to modify the main.go file. Although this file probably shouldn't live in your GOPATH, feel free to import libraries from your GOPATH (including libraries downloaded with go get).

## Local Testing
Run ```make test``` to compile your code and start the test server. Open ```http://localhost:8080/req``` in your browser. The page should display ```User function is ready```. Refresh the page to talk to your code.

## Deployment
Run ```make``` to compile and package your code. Upload the generated ```function.zip``` file to Google Cloud Functions as an HTTP trigger function.

## Limitations
* This has only been tested for HTTP trigger functions. Non-HTTP trigger functions will use a different endpoint (not ```/req```).
* Logging is not supported (yet).

## Troubleshooting
Some versions of Node.js (especially those packaged for Ubuntu) name their main binary ```nodejs``` instead of ```node```. The symptom of this problem is an error about the ```node``` binary not being found in the path even though Node.js is installed. This can be fixed with ```sudo ln -s $(which nodejs) /usr/local/bin/node```.

## License

Copyright 2017, Google, Inc.

Licensed under the Apache License, Version 2.0

See [LICENSE](LICENSE).
