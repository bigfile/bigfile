# Bigfile —— Manage Files In A Different Way

<img align="right" width="159px" src="https://avatars3.githubusercontent.com/u/52916753">

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/bigfile/bigfile)](https://goreportcard.com/report/github.com/bigfile/bigfile)
[![Open Source Helpers](https://www.codetriage.com/bigfile/bigfile/badges/users.svg)](https://www.codetriage.com/bigfile/bigfile)

Bigfile is built on top of many excellent open source projects. Designed to provide a file management service and give developers more help. At the bottom, bigfile splits the file into small pieces of 1MB, only the last shard may be less than 1mb, the same slice will only be stored once.

We also built a virtual file organization system that is logically divided into directories and files. You can delete directories, files, and move. You can also use the ftp service we provide to manage files. That's just too cool. In the development project, you only need to use the http interface to access, in the future we will also develop various sdk, reduce the difficulty of use.


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_large)
