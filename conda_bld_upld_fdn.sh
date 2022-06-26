#!/bin/bash

shopt -s expand_aliases

if ! command -v conda &>/dev/null; then
  alias conda=/usr/local/anaconda3/bin/conda
  if ! command -v conda &>/dev/null; then
    echo "conda not exist"
    exit 127
  fi
fi

if ! command -v anaconda &>/dev/null; then
  alias anaconda=/usr/local/anaconda3/bin/anaconda
  if ! command -v anaconda &>/dev/null; then
    echo "anaconda not exist"
    exit 127
  fi
fi

# set Python versions
platform=$(conda info | grep -i platform | cut -d ":" -f 2 | xargs)
pkg="fdn"
array=(3.8 3.9 3.10)
printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf "Building conda package ..."

# create package version and number
#ver=$(date '+%Y.%m.%d')
#num=$(date '+%H%M')
#
#sed -i "s/.*{% set ver = \".*\" %}.*/{% set ver = \"$ver\" %}/" meta.yaml
#sed -i "s/.*{% set num = \".*\" %}.*/{% set num = \"$num\" %}/" meta.yaml
#sed -i "s/.*_ver.*=.*\".*\".*/_ver = \"$ver.$num\"/" fdn/fdnlib/fdncli.py

# building conda packages
for i in "${array[@]}"; do
  conda build . --python "$i"
done

# clear setting
#sed -i "s/.*{% set ver = \".*\" %}.*/{% set ver = \"XXXX.XX.XX\" %}/" meta.yaml
#sed -i "s/.*{% set num = \".*\" %}.*/{% set num = \"XXXX\" %}/" meta.yaml
#sed -i "s/.*_ver.*=.*\".*\".*/_ver = \"XXXX.XX.XX\"/" fdn/fdnlib/fdncli.py

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
read -p "Convert to others platform? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
  printf "Convert to other platform ...\n"
  bld_dir=$HOME/anaconda3/conda-bld

  # convert package to other platforms
  platforms=(osx-64 osx-arm64 linux-32 linux-64 win-32 win-64)

  # shellcheck disable=SC2061
  # shellcheck disable=SC2154
  # shellcheck disable=SC2162
  find "$bld_dir"/"$platform"/ -name $pkg*"$ver"*"$num"*.tar.bz2 | while read file; do
    # shellcheck disable=SC2059
    printf "$file"
    #conda convert --platform all $file  -o $HOME/conda-bld/
    for platform in "${platforms[@]}"; do
      # shellcheck disable=SC2086
      conda convert --platform "$platform" $file -o $bld_dir/
    done

  done
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
  printf "Convert finished.\n"
else
  exit 1
fi

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf 'Please check anaconda current user information'
printf '\n'
printf '********************************************'
printf '\n'
anaconda whoami
printf '********************************************'
printf '\n'
read -p "User information is right? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Nn][Oo]$ ]]; then
  anaconda login
fi
###############################################################################
read -p "Upload to anaconda? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
  # upload packages to conda
  # shellcheck disable=SC2061
  # shellcheck disable=SC2162
  find "$bld_dir"/ -name $pkg*"$ver"*"$num"*.tar.bz2 | while read file; do
    # shellcheck disable=SC2059
    printf "$file"
    anaconda upload "$file"
  done
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
  printf "Upload finished.\n"
else
  exit 1
fi

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
read -p "Purge all and Delete all local packages? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
  conda build purge-all
  # shellcheck disable=SC2061
  # shellcheck disable=SC2162
  find "$bld_dir"/ -name $pkg*.tar.bz2* | while read file; do
    # shellcheck disable=SC2059
    printf "$file"
    printf "\n"
    rm "$file"
  done
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
  printf "All packages deleted.\n"
else
  exit 1
fi
