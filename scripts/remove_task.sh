#!/bin/bash

url="http://127.0.0.1:7117/peck_task/remove"
config='{"Name":"TestLog","LogPath":".test.log"}'

curl -XPOST $url -d $config