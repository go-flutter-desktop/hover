#!/bin/bash

function finish {
	# If we don't know what UID/GID the user outside docker has, we cannot fix
	# file ownership. This seems to be fine when on windows.
	if [[ ${HOVER_SAFE_CHOWN_UID} && ${HOVER_SAFE_CHOWN_UID-x} ]]; then
		chown -R ${HOVER_SAFE_CHOWN_UID}:${HOVER_SAFE_CHOWN_GID} /app
		chown -R ${HOVER_SAFE_CHOWN_UID}:${HOVER_SAFE_CHOWN_GID} /root/.cache/hover
		chown -R ${HOVER_SAFE_CHOWN_UID}:${HOVER_SAFE_CHOWN_GID} /go-cache
		# echo "chowned files to ${HOVER_SAFE_CHOWN_UID}:${HOVER_SAFE_CHOWN_GID}"
	fi
}
trap finish EXIT

hover $@
