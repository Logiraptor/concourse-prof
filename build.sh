#!/bin/bash

pushd frontend
  yarn build
popd

zip -r package.zip *.go go.mod go.sum frontend/public/* frontend/build/*

