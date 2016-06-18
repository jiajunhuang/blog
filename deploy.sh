#!/bin/bash

git commit -m "$@"
git push
cd ../
git commit -m "submodule && $@"
git push
