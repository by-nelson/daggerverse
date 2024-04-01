#!/bin/sh

# This function chain is going to format the test.py file
dagger call with-version with-yapf with-directory --directory "$(pwd)" format --apply get-directory export --path "$(pwd)" 
