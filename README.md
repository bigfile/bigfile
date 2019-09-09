### Bigfile

<img align="right" width="159px" src="https://avatars3.githubusercontent.com/u/52916753">

[![Build Status](https://travis-ci.org/bigfile/bigfile.svg?branch=master)](https://travis-ci.org/bigfile/bigfile)
[![codecov](https://codecov.io/gh/bigfile/bigfile/branch/master/graph/badge.svg)](https://codecov.io/gh/bigfile/bigfile)
[![GoDoc](https://godoc.org/github.com/bigfile/bigfile?status.svg)](https://github.com/bigfile/bigfile)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/bigfile/bigfile)](https://goreportcard.com/report/github.com/bigfile/bigfile)
[![Open Source Helpers](https://www.codetriage.com/bigfile/bigfile/badges/users.svg)](https://www.codetriage.com/bigfile/bigfile)

**Bigfile** is a file transfer system, supports http, ftp and rpc protocol. It is built on top of many excellent open source projects. Designed to provide a file management service and give developers more help. At the bottom, bigfile splits the file into small pieces of 1MB, the same slice will only be stored once. Please allow me to illustrate the entire architecture with a picture.

![architecture](https://bigfile.site/bigfile.png)

In fact, we built a file organization system based on the database. Here you can find familiar files and folders. But in Bigfile, files and folders both are considered to be files.

Since the rpc and http protocols are supported, those languages supported by [grpc](https://grpc.io/) and other languages can be quickly accessed. If you are not a programmer, you can use the ftp client to manage your files, the only thing you need to do is start Bigfile.

You can find more detailed [Documentation](https://bigfile.site) here. **Windows platform is currently unavailable**.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_large)