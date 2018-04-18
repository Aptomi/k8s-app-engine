#!/bin/bash

fly -t tutorial login
fly -t tutorial sync
fly -t tutorial destroy-pipeline -n -p dev-to-prod
