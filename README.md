<p align="center">
    ![architecture](https://bigfile.site/bigfile.png)
</p>

<p align="center">
    [![Build Status](https://travis-ci.org/bigfile/bigfile.svg?branch=master)](https://travis-ci.org/bigfile/bigfile)
    [![codecov](https://codecov.io/gh/bigfile/bigfile/branch/master/graph/badge.svg)](https://codecov.io/gh/bigfile/bigfile)
    [![GoDoc](https://godoc.org/github.com/bigfile/bigfile?status.svg)](https://github.com/bigfile/bigfile)
    [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_shield)
    [![Go Report Card](https://goreportcard.com/badge/github.com/bigfile/bigfile)](https://goreportcard.com/report/github.com/bigfile/bigfile)
    [![Open Source Helpers](https://www.codetriage.com/bigfile/bigfile/badges/users.svg)](https://www.codetriage.com/bigfile/bigfile)
    [![中文文档](https://img.shields.io/badge/%E6%96%87%E6%A1%A3-%E4%B8%AD%E6%96%87-blue)(https://learnku.com/docs/bigfile)
    [![English Documentation](https://img.shields.io/badge/Doc-English-blue)(https://bigfile.site)
</p>

**Bigfile** is a file transfer system, supports http, ftp and rpc protocol. Designed to provide a file management service and give developers more help. At the bottom, bigfile splits the file into small pieces of 1MB, the same slice will only be stored once.In fact, we built a file organization system based on the database. Here you can find familiar files and folders. But in Bigfile, files and folders both are considered to be files.Since the rpc and http protocols are supported, those languages supported by [grpc](https://grpc.io/) and other languages can be quickly accessed. If you are not a programmer, you can use the ftp client to manage your files, the only thing you need to do is start Bigfile.

----

# Features

* Support HTTP(s) protocol

    * Support rate limit by ip
    * Support cors
    * Support to avoid replay attack
    * Support to validate parameter signature

* Support FTP(s) protocol

* Support RPC protocol

* Support deploy by docker

* Provide document with English and Chinese

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_large)
