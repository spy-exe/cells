#!/usr/bin/env bash
set -euo pipefail
# Execute on the personal pve host only. Refuse existing resources.
[[ "$(hostname)" == pve ]]
ctid=${CELLS_CTID:-220}
[[ "$ctid" == 220 ]]
[[ "$(pvesh get /cluster/nextid --vmid "$ctid")" == "$ctid" ]]
[[ ! -e /etc/pve/lxc/$ctid.conf ]]
[[ $(df --output=avail -k / | tail -1) -gt 2097152 ]]
template=local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst
pveam list local | grep -F "$template" >/dev/null
if nmap -sn -PR 10.10.1.220 | grep -q 'Host is up'; then
  echo 'Refusing occupied address 10.10.1.220' >&2; exit 1
fi
pct create "$ctid" "$template" --hostname cells-prd-01 --unprivileged 1 \
  --cores 4 --cpulimit 2 --memory 6144 --swap 512 --rootfs local-lvm:40 \
  --net0 name=eth0,bridge=vmbr0,ip=10.10.1.220/24,gw=10.10.1.254,type=veth \
  --nameserver 1.1.1.1 --timezone America/Sao_Paulo --onboot 1 \
  --description 'Cells multiplayer — dedicated personal project; native systemd services; no nesting'
pct start "$ctid"
