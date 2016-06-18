#!/bin/bash

git commit -m "$@"
git push
cd ../
git commit -am "submodule && $@"
git push
