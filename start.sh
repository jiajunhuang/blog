#!/bin/bash

rm -f assets.go && go-assets-builder templates/*.tpl -o asserts.go && go build && ./blog
