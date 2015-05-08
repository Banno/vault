#!/bin/bash
TF_ACC=yes MARATHON_HOSTNAME="dev.banno.com" go test . -v
