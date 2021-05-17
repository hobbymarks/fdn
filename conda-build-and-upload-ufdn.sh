#!/bin/bash

# the Python versions
array=(3.8 3.9)

echo "Building conda package ..."

# create version and number
ver=$(date '+%Y.%m.%d')
num=$(date '+%H%M')

sed -i "1s/.*/{% set ver = \"$ver\" %}/" meta.yaml
sed -i "2s/.*/{% set ver = \"$num\" %}/" meta.yaml

# building conda packages
for i in "${array[@]}"; do
  conda build . --python $i
done

## convert package to other platforms
#cd ~
#platforms=(osx-64 linux-32 linux-64 win-32 win-64)
#find $HOME/conda-bld/linux-64/ -name *.tar.bz2 | while read file; do
#  echo $file
#  #conda convert --platform all $file  -o $HOME/conda-bld/
#  for platform in "${platforms[@]}"; do
#    conda convert --platform $platform $file -o $HOME/conda-bld/
#  done
#
#done
#
## upload packages to conda
#find $HOME/conda-bld/ -name *.tar.bz2 | while read file; do
#  echo $file
#  anaconda upload $file
#done
#
#echo "Building conda package done!"
