#!/bin/bash

python gen_catalog.py

git add .

git commit $@
