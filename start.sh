#!/bin/bash

gunicorn -c gunicorn_config.py web:app
