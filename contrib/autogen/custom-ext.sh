#!/bin/bash
# Author: Daniele Rondina, geaaru@macaronios.org

process() {
  local f=$1

  cd $downloaddir

  # Create tarball and/or single file on download dir
  # as local artefact.

  echo "mydata" > ${downloaddir}/${name}-${version}

  echo "
artefacts:
  - url: ${mirror}/${name}-${version}
    name: ${name}-${version}
    local: true
" > $f

  return 0
}

read_vars() {
  local f=$1

  #cat $f

  # Read atom information
  name=$(yq4 e '.name' $f)
  version=$(yq4 e '.vars.version' $f)
  mirror=$(yq4 e '.vars.mirror' $f)
  category=$(yq4 e '.atom.category' $f)
  downloaddir=$(yq4 e '.vars.download_dir' $f)

  #echo "Elaborating package ${name}-${version} of category ${category} on download dir ${downloaddir}..."

  export name category downloaddir version mirror

  return 0
}

main() {
  local source_file=$1
  local target_file=$2

  read_vars "${source_file}" || return 1
  process "${target_file}" || return 1

  return 0
}

main $@
exit $?

