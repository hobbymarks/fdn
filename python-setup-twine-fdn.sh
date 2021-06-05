#!/bin/bash

printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf "Building pypi package ..."
python setup.py bdist_wheel sdist


printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
printf "Twine checking ..."
twine check dist/*

printf '%*s\n\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
###############################################################################
read -p "upload to pypi platform? yes/[no] " -r
printf '\n'
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
  twine upload --repository pypi dist/*
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' \#
  printf "upload finished.\n"
else
  exit 1
fi

