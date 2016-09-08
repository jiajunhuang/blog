#!/bin/bash

# generate catalog
/usr/bin/python3 gen_catalog.py

if [ $# -eq 0  ]; then
    git commit -m "new post"
else
    git commit -m "$1"
fi

git push
