#!/bin/bash
# Author: Daniele Rondina, geaaru@macaronios.org
# Description:

process() {
  local f=$1

  echo "
vars:
  versions:
    - '1.5.0'
    - '1.6.0'
    - '1.7.0'
artefacts:
  - url: https://www.x.org/releases/individual/lib/libX11-1.7.0.tar.gz
    name: libX11-1.7.0.tar.gz
" > $f

  echo "File $f written"
  cat $f

  return 0
}

set_version() {

  return 0
}

read_vars() {
  local f=$1

  cat $f

  # Read atom information
  local name=$(yq4 e '.name' $f)
  local category=$(yq4 e '.atom.category' $f)

  echo "Elaborating package ${name} of category ${category}..."

  export name category

  return 0
}

main() {
  local mode=$1
  local source_file=$2
  local target_file=$3

  case $mode in
    process)
      read_vars "${source_file}" || return 1
      process "${target_file}" || return 1
      ;;

    set-version)
      read_vars "${source_file}" || return 1
      ;;

    *)
      echo "Unsupported mode ${mode}!"
      return 1
      ;;

  esac

  return 0
}

main $@
exit $?
