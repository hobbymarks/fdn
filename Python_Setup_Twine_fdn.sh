#!/bin/bash

printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf "Clean Projec ...\n"
p='fdn/data/corpora/words'
if [ -d $p ]; then
  printf "delete $p\n"
  rm -rf "$p"
fi
p='build'
if [ -d $p ]; then
  printf "delete $p\n"
  rm -rf "$p"
fi
p='dist'
if [ -d $p ]; then
  printf "delete $p\n"
  rm -rf "$p"
fi
p='fdn.egg-info'
if [ -d $p ]; then
  printf "delete $p\n"
  rm -rf "$p"
fi

printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf "Building pypi package ..."
python setup.py bdist_wheel sdist

printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf "Twine checking ..."
if ! command -v twine &>/dev/null; then
  echo "twine(https://github.com/pypa/twine) not exist."
  exit 127
fi
twine check dist/*

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
read -p "upload to pypi platform? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
  twine upload --repository pypi dist/*
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
  printf "upload finished.\n"
fi

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
read -p "clean all build files in local? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
  rm -fr dist
  rm -fr build
  rm -fr fdn.egg-info
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
  printf "clean finished.\n"
else
  exit 1
fi
