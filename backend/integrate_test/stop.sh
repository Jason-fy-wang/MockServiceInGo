#!/usr/bin/env bash


ps -ef | grep starter | grep -v grep | awk '{print $2}' | xargs kill 