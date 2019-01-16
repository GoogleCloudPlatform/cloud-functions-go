# Unofficial Native Go Runtime for Google Cloud Functions

_Disclaimer: This is not an official Google product. It is not and will not be maintained by Google, and is not part of Google Cloud Functions project. There is no guarantee of any kind, including that it will work or continue to work, or that it will supported in any way._

## Looking for the official Go Runtime for Google Cloud Functions?
Google Cloud Functions supports an official Go runtime. Check out the [quickstart](https://cloud.google.com/functions/docs/quickstart#functions-update-install-gcloud-go) to try it out.

## Background
When this doc was first written, Google Cloud Functions only supported a Node.js runtime. Since then, new official runtimes have been added. Using an official runtime (such as the official Go runtime) is the recommended approach for almost all users. 

But if you're feeling adventurous, you can run non-JavaScript code by using native Node modules and running subprocesses. Node will always be handling the requests. Wrapping Go code in a native Node module will require writing complicated foreign function interfaces (FFIs). FFIs are notoriously difficult to get right and are an easy way to introduce memory leaks as well as security and stability issues. Using a subprocess is much easier to get right, but it introduces the overhead of interprocess communication (IPC) and context switches. In both cases, the requests are proxied between Node and Go which typically involves extra copies and translations.

## 100% Pure Go Request Path
After an initial bootstrap step, the only thing running is a pure Go binary. This project works by completely replacing the Node process and lets a Go process directly handle requests. There is no cgo and no proxying. This means that it is faster and potentially more secure than other projects which typically to use both.

## How It Works
A native module calls execve(2) without the normal preceding fork(2) call when the user module is imported. Open socket FDs are preserved and handled by the new Go process to ensure a smooth transition. The new Go process then pretends to be Node.

## Requirements
There are four supported environments:

### Native
For advanced users who don't like VirtualBox.
* Linux or macOS development environment (The native Node module depends on Linux/macOS specific behavior.)
* Go 1.5 or above
* Node.js v0.10 or above, npm and node-gyp
* Make and GCC
* If you are using macOS then as of right now you need the FULL xcode installed. Command line tools alone will not work.

### Vagrant
Compatible with Windows, macOS and Linux and easier than native development.
* [Vagrant](https://www.vagrantup.com/downloads.html)
* [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or a compatable provider
  * Optional: [vagrant-vbguest](https://github.com/dotless-de/vagrant-vbguest)

In the `cloud-functions-go` directory, run `vagrant up` to start the environment and `vagrant ssh` to get a shell. Support for other Vagrant providers besides VirtualBox is in the works.

### Windows with Cygwin
* Go for Windows
* Node.js for Windows
* [Cygwin](https://cygwin.com/install.html) with `make` and `zip` (you may also want `git` and an editor like `vim` or `nano`)

Use the Cygwin Terminal to run the commands as described below. Note that `make test` won't work under Windows.

### Windows with Powershell 5.0
* Go 1.5 or above
* Node.js (optional)

The commands described below may be run as-is using Command Prompt, or prefixed with `./` using Windows PowerShell (i.e. `./make` or `./make godev`). Note that `make test` won't work under Windows.

### External Dependencies

The `events` sub-package depends on the following libraries:
* `google.golang.org/api/pubsub/v1`
* `google.golang.org/api/storage/v1`

## Hello, world!
A demo hello world example is included. To try it out, simply skip to the [Deployment](#deployment) section.

## Making Changes
For normal development, it should only be necessary to modify the main.go file. Although this file probably shouldn't live in your GOPATH, feel free to import libraries from your GOPATH (including libraries downloaded with go get).

## Local Testing
Run ```make test``` to compile your code and start the test server. Open ```http://localhost:8080/execute``` in your browser. The page should display ```User function is ready```. Refresh the page to talk to your code.

When using a go-only or non-unix environment, run ```make godev``` instead.

## Logging
The logger may be used directly:
```
nodego.InfoLogger.Println("Hello World!")
nodego.ErrorLogger.Println("Something went wrong!")
```
or via the log package after calling `nodego.OverrideLogger()`:
```
func init() {
	nodego.OverrideLogger()
}

...

log.Println("Hello World!")
```
Use the `nodego.WithLogger()` or `nodego.WithLoggerFunc()` middleware for your logs to reference the specific execution:
```
http.WithLoggerFunc(func(w http.ResponseWriter, r *http.Request) {
	log.Println("Hello World!")
})
```
A full example is included in [examples/logging.go](examples/logging.go).

## Deployment
Run ```make``` to compile and package your code. Upload the generated ```function.zip``` file to Google Cloud Functions as an HTTP trigger function.

### Vagrant
Run ```vagrant up``` to start the envirement. Run ```vagrant ssh``` to connect to the envirement. Run ```cd /vagrant``` to access the respority files. The instructions in [Local Testing](#local-testing) and [Deployment](#deployment) should now work.

## Troubleshooting
Some versions of Node.js (especially those packaged for Ubuntu) name their main binary ```nodejs``` instead of ```node```. The symptom of this problem is an error about the ```node``` binary not being found in the path even though Node.js is installed. This can be fixed with ```sudo ln -s $(which nodejs) /usr/local/bin/node```. There's also a package called `nodejs-legacy` that can be installed in some Debian and Ubuntu distros using `apt` that creates a symlink `node` in `/usr/bin/`

It may also be helpful to understand the nodejs environment that your Google Cloud Function runs in (e.g. how does regular `console.log` send logs to Stackdriver, whereas `fmt.Println` doesn't?). You can find more details in the [WORKERJS.md](WORKERJS.md) file.

## License
Copyright 2017, Google, Inc.

Licensed under the Apache License, Version 2.0

See [LICENSE](LICENSE).
