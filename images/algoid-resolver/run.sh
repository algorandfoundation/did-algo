#!/bin/sh

# Automatically exit on error
set -e

# Use 'exec' so that the 'algoid' application becomes the container’s
# PID 1. This allows the application to receive any Unix signals sent
#to the container.
exec algoid resolver
