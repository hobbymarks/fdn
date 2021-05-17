#!/bin/bash

# the Python versions
pkg="ufdn"
array=(3.8 3.9)
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
echo "Building conda package ..."

# create version and number
ver=$(date '+%Y.%m.%d')
num=$(date '+%H%M')

sed -i "1s/.*/{% set ver = \"$ver\" %}/" meta.yaml
sed -i "2s/.*/{% set num = \"$num\" %}/" meta.yaml

# building conda packages
for i in "${array[@]}"; do
  conda build . --python $i
done

sed -i "1s/.*/{% set ver = \"XXXX.XX.XX\" %}/" meta.yaml
sed -i "2s/.*/{% set num = \"XXXX\" %}/" meta.yaml
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
read -p "Convert to others platform? [y]/n" -n 1 -r
echo # (optional) move to a new line
if [[ $REPLY =~ ^[Nn]$ ]]; then
  exit 1
fi
echo "Convert conda package to other platform ..."
bld_dir=$HOME/anaconda3/conda-bld

# convert package to other platforms
platforms=(osx-64 osx-arm64 linux-32 linux-64 win-32 win-64)

find $bld_dir/linux-64/ -name $pkg*$ver*$num*.tar.bz2 | while read file; do
  echo $file
  #conda convert --platform all $file  -o $HOME/conda-bld/
  for platform in "${platforms[@]}"; do
    conda convert --platform $platform $file -o $bld_dir/
  done

done
echo "Convert conda package finished.\n\n\n"
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
read -p "Upload to anaconda? [y]/n" -n 1 -r
echo # (optional) move to a new line
if [[ $REPLY =~ ^[Nn]$ ]]; then
  exit 1
fi

# upload packages to conda
find $bld_dir/ -name $pkg*$ver*$num*.tar.bz2 | while read file; do
  echo $file
  anaconda upload $file
done
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
echo "Upload finished."
