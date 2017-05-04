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

#include <dirent.h>
#include <errno.h>
#include <fcntl.h>
#include <node.h>
#include <stdio.h>
#include <stdlib.h>
#include <string>
#include <sys/socket.h>
#include <unistd.h>
#include <v8.h>
#include <vector>

using namespace v8;

void init(Handle<Object> target) {
	const char bin[] = "./main";

	// Clear CLOEXEC for STDOUT and STDERR.
	if (fcntl(STDOUT_FILENO, F_SETFD, 0) == -1) {
		fprintf(stderr, "fcntl(STDOUT_FILENO, F_SETFD, 0) %d\n", errno);
		exit(1);
	}
	if (fcntl(STDERR_FILENO, F_SETFD, 0) == -1) {
		fprintf(stderr, "fcntl(STDERR_FILENO, F_SETFD, 0) %d\n", errno);
		exit(1);
	}

	DIR *dir = opendir(
		#ifdef __APPLE__
		"/dev/fd"
		#else
		"/proc/self/fd"
		#endif
	);
	if (dir == NULL) {
		fprintf(stderr, "opendir failed\n");
		exit(1);
	}

	std::vector<std::string> fds;

	for (struct dirent *ent = readdir(dir); ent != NULL; ent = readdir(dir)) {
		int fd = atoi(ent->d_name);
		struct sockaddr_storage addr;
		socklen_t addrlen = sizeof(addr);
		if (getsockname(fd, reinterpret_cast<sockaddr*>(&addr), &addrlen) >= 0) {
			// Clear CLOEXEC for the socket FD.
			if (fcntl(fd, F_SETFD, 0) == -1) {
				fprintf(stderr, "fcntl(%d, F_SETFD, 0) %d\n", fd, errno);
				exit(1);
			}
			fds.push_back(ent->d_name);
		}
	}

	std::vector<const char*> args;
	args.push_back(bin);
	for (size_t i = 0; i < fds.size(); ++i) {
		args.push_back(fds[i].c_str());
	}
	args.push_back(NULL);

	execv(bin, const_cast<char* const*>(&args[0]));

	fprintf(stderr, "execve %d\n", errno);
	exit(1);
}
NODE_MODULE(execer, init)
