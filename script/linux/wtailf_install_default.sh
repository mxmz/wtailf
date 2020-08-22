#!/bin/bash

set -Eeuo pipefail

exeurl="https://gitlab.com/mxmz/wtailf/-/jobs/698899335/artifacts/raw/go/cmd/wtailf/wtailf"
sources="/var/log /home/*/var/log /var/log/syslog /var/log/messages"
servicename="wtailf"
installdir=/opt/wtailf
main() {



    mkdir -p "$installdir/bin"
#   mkdir -p ~/.config/systemd/user/

    cd "$installdir"
    curl -O $exeurl
    chmod +x ./wtailf
    mv -v ./wtailf bin/

    exepath=$installdir/bin/wtailf

    dirs=

    for d in $sources; do
        if [[ -d $d ]] || [[ -f $d ]] ; then
            dirs="$dirs $d"
        fi
    done

    cat << EOF > /lib/systemd/system/$servicename.service	
[Unit]
Description=WTailF
After=network.target nss-lookup.target

[Service]
ExecStart=$exepath ":18081" $dirs
TimeoutStopSec=5
KillMode=process
User=syslog

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
	systemctl enable wtailf
	systemctl restart wtailf



}



main "$@"
