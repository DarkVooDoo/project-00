#!/bin/bash

ffmpeg -i $1 -vf scale=400:-1 -y test.jpg
