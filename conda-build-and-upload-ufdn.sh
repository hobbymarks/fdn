#!/bin/bash

# set Python versions
pkg="ufdn"
array=(3.8 3.9)
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
printf "Building conda package ..."

# create package version and number
ver=$(date '+%Y.%m.%d')
num=$(date '+%H%M')

sed -i "s/.*{% set ver = \".*\" %}.*/{% set ver = \"$ver\" %}/" meta.yaml
sed -i "s/.*{% set num = \".*\" %}.*/{% set num = \"$num\" %}/" meta.yaml
sed -i "s/.*_ver.*=.*\".*\".*/_ver = \"$ver.$num\"/" ufdn/ufdnlib/ufncli.py

# building conda packages
for i in "${array[@]}"; do
  conda build . --python "$i"
done

# clear setting
sed -i "s/.*{% set ver = \".*\" %}.*/{% set ver = \"XXXX.XX.XX\" %}/" meta.yaml
sed -i "s/.*{% set num = \".*\" %}.*/{% set num = \"XXXX\" %}/" meta.yaml
sed -i "s/.*_ver.*=.*\".*\".*/_ver = \"XXXX.XX.XX\"/" ufdn/ufdnlib/ufncli.py

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -

read -p "Convert to others platform? [y]/n" -n 1 -r
printf '\n' # (optional) move to a new line
if [[ $REPLY =~ ^[Nn]$ ]]; then
  exit 1
fi
printf "Convert to other platform ..."
bld_dir=$HOME/anaconda3/conda-bld

# convert package to other platforms
platforms=(osx-64 osx-arm64 linux-32 linux-64 win-32 win-64)

find $bld_dir/linux-64/ -name $pkg*$ver*$num*.tar.bz2 | while read file; do
  printf $file
  #conda convert --platform all $file  -o $HOME/conda-bld/
  for platform in "${platforms[@]}"; do
    conda convert --platform $platform $file -o $bld_dir/
  done

done
printf "Convert finished.\n"

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
read -p "Upload to anaconda? [y]/n" -n 1 -r
printf # (optional) move to a new line
if [[ $REPLY =~ ^[Nn]$ ]]; then
  exit 1
fi

# upload packages to conda
find $bld_dir/ -name $pkg*$ver*$num*.tar.bz2 | while read file; do
  printf $file
  anaconda upload $file
done
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
printf "Upload finished."
