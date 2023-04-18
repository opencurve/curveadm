#!/usr/bin/env bash

# Created Date: 2021-12-27
# Author: Jingli Chen (Wine93)


############################  GLOBAL VARIABLES
g_curve_format="{{.curve_format}}"
g_percent="{{.percent}}"
g_chunkfile_size="{{.chunkfile_size}}"
g_chunkfile_pool_dir="{{.chunkfile_pool_dir}}"
g_chunkfile_pool_meta_path="{{.chunkfile_pool_meta_path}}"

############################ FUNCTIONS
main() {
    mkdir -p "${g_chunkfile_pool_dir}"
    "${g_curve_format}" \
      -allocatePercent="${g_percent}" \
      -fileSize=${g_chunkfile_size} \
      -filePoolDir="${g_chunkfile_pool_dir}" \
      -filePoolMetaPath=${g_chunkfile_pool_meta_path} \
      -fileSystemPath=${g_chunkfile_pool_dir}
}

############################  MAIN()
main "$@"
