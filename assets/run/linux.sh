#!/bin/sh

path=`dirname "${0}"`
link=`readlink -f "${0}"`

[ -n "${link}" ] && path=`dirname "${link}"`
cd "${path}"

./{{.}} "${@}"
