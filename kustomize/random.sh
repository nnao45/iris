#!/bin/sh

echo iris-$(LC_CTYPE=C tr -dc A-Za-z0-9 < /dev/urandom | fold -w ${1:-16} | head -n 1)
