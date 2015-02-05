Shipmate
========

Utility command for executing multiple docker builds in a repository.

Installation
------------

Assuming you have a proper golang environment setup:

```
go get github.com/oxtopus/shipmate
go install github.com/oxtopus/shipmate
```

Usage
-----

```
$ shipmate -h
Usage of shipmate:
  -name="": Local destination of bare repository
  -remote="": Remote repository URL
  -rev="master": Git revision
```

`shipmate` assumes that it is run in the context of a directory that contains
one or more Dockerfiles (recursively) that permute multiple linux
environments. For example, consider a directory tree that looks like this:

```
.
├── centos
│   ├── 6
│   │   └── gcc
│   │       └── Dockerfile
│   └── 7
│       └── gcc
│           └── Dockerfile
├── debian
│   └── jessie
│       └── clang
│           └── Dockerfile
└── ubuntu
    └── 14.04
        ├── clang
        │   └── Dockerfile
        └── gcc
            └── Dockerfile
```

Followed by this command:

```
shipmate -name=$NAME -remote=$REMOTE -rev=$REV
```

The `shipmate` process will clone the remote repository defined by `$REMOTE`,
reset to the revision `$REV`, and then recursively search the current working
directory for Dockerfiles, excluding what you just cloned.  `shipmate` will
make a shallow clone relative to the directories that it finds and execute a
`docker build` command.  The docker repository will be assumed to be `$NAME`,
and the tag constructed according to the `$REV` value and the path to the
Dockerfile being built.  The final result will be docker images in the `$NAME`
repository for each of the Dockerfiles that were found, with unique tags of
the format `$REV-SUFFIX`, where `SUFFIX` is a translation of the path to use
dashes instead of slashes, e.g. `debian-jessie-clang` or `ubuntu-14.04-gcc`
citing the example above.

License
-------

Copyright (c) 2015, Austin Marshall <oxtopus@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the distribution.
3. All advertising materials mentioning features or use of this software
   must display the following acknowledgement:
   This product includes software developed by Austin Marshall.
4. Neither the name "Shipmate" nor the
   names of its contributors may be used to endorse or promote products
   derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY AUSTIN MARSHALL ''AS IS'' AND ANY
EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL AUSTIN MARSHALL BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
