#!/bin/sh
set -e

git status
git add .
git commit -m "submit local update for release"
git status

curl -sL https://git.io/goreleaser | bash