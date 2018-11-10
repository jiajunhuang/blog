#!/bin/bash

gunicorn -c gunicorn_config.py main:app
