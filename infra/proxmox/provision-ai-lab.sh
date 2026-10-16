#!/usr/bin/env bash
# Project-owned, disposable CPU lab. Never run this on a corporate node.
set -euo pipefail
[[ "$(hostname)" == pve ]]
[[ "$(pvesh get /cluster/nextid --vmid 221)" == 221 ]]
[[ ! -e /etc/pve/lxc/221.conf ]]
[[ $(awk '/MemAvailable/{print $2}' /proc/meminfo) -gt 3145728 ]]
if nmap -sn -PR 10.10.1.221 | grep -q 'Host is up'; then
  echo 'Address 10.10.1.221 is occupied' >&2; exit 1
fi
pct create 221 local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst \
 --hostname cells-ai-lab-01 --unprivileged 1 --cores 1 --cpulimit 0.75 \
 --memory 1536 --swap 256 --rootfs local-lvm:8 --onboot 0 \
 --net0 name=eth0,bridge=vmbr0,ip=10.10.1.221/24,gw=10.10.1.254,type=veth \
 --nameserver 1.1.1.1 --timezone America/Sao_Paulo \
 --description 'Cells isolated CPU training lab. Stop when idle; no production data, no nesting.'
pct start 221
pct exec 221 -- bash -c 'set -e; chmod 1777 /tmp /run/lock; systemctl mask tmp.mount run-lock.mount dev-mqueue.mount; systemctl reset-failed; apt-get update -qq; DEBIAN_FRONTEND=noninteractive apt-get install -y -qq python3-venv ca-certificates curl nftables; useradd --system --create-home --home-dir /var/lib/cells-ai --shell /usr/sbin/nologin cells-ai; mkdir -p /opt/cells-ai/src /var/lib/cells-ai/runs; chown -R cells-ai:cells-ai /var/lib/cells-ai; python3 -m venv /opt/cells-ai/venv'
pct exec 221 -- systemctl disable --now ssh.service ssh.socket
pct exec 221 -- bash -c 'set -e; cat > /etc/nftables.conf <<"NFT"
#!/usr/sbin/nft -f
table inet cells_lab {
  chain input {
    type filter hook input priority 0; policy drop;
    iifname "lo" accept
    ct state established,related accept
    ip protocol icmp accept
    meta l4proto ipv6-icmp accept
  }
  chain forward { type filter hook forward priority 0; policy drop; }
  chain output { type filter hook output priority 0; policy accept; }
}
NFT
nft -c -f /etc/nftables.conf; systemctl enable --now nftables'
