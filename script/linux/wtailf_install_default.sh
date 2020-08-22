#!/bin/bash
set -Eeuo pipefail

: ${WTAILF_BUILDID:="698921935"}
: ${WTAILF_SOURCES:="/var/log /home/*/var/log /var/log/syslog /var/log/messages"}
: ${WTAILF_USER:="wtailf"}
: ${WTAILF_LISTEN:=":18081"}


exeurl="https://gitlab.com/mxmz/wtailf/-/jobs/$WTAILF_BUILDID/artifacts/raw/go/cmd/wtailf/wtailf"
servicename="wtailf"
installdir=/opt/wtailf



main() {

    mkdir -p "$installdir/bin"

    cd "$installdir"
    curl -O $exeurl
    chmod +x ./wtailf
    mv -v ./wtailf bin/

    exepath=$installdir/bin/wtailf

    dirs=

    for d in $WTAILF_SOURCES; do
        if [[ -d $d ]] || [[ -f $d ]] ; then
            dirs="$dirs $d"
        fi
    done

    cat << EOF > /lib/systemd/system/$servicename.service	
[Unit]
Description=WTailF [https://gitlab.com/mxmz/wtailf/-/jobs/$WTAILF_BUILDID]
After=network.target nss-lookup.target

[Service]
ExecStart=$exepath "$WTAILF_LISTEN" $dirs
TimeoutStopSec=5
KillMode=process
User=$WTAILF_USER

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
	systemctl enable wtailf
	systemctl restart wtailf



}



main "$@"
