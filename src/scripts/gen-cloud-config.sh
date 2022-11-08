#!/bin/bash

v_CLOUD_CONFIG="$1"

echo -e ${v_CLOUD_CONFIG} | sudo tee cloud-config
