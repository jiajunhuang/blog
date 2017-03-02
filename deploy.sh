#!/bin/bash

python gen_catalog.py

git add .

git commit -m "new post on `date`"

`command -v proxychains 2>&1` git push
