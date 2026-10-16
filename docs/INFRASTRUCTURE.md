# Descoberta e limites — 16/10/2026

Escopo: homelab pessoal de spy-exe, nó único `pve` (Proxmox 9.2.4, Ryzen 7 3700U, 8 CPUs lógicas, 17,5 GiB RAM). Nenhum acesso a ambientes corporativos faz parte deste projeto.

Observado antes de provisionar: ~6,6 GiB de memória disponível; swap 6,7/8 GiB ocupada. `local-lvm` thin pool 156,87 GiB, 65,21% usado (~54,6 GiB disponíveis), metadados 3,58%. Filesystem do host tem somente ~3 GiB disponíveis: não instalar toolchains, modelos ou backups grandes nele. CTs/VMs existentes preservados.

Rede: vmbr0 em 10.10.1.0/24, sem VLAN, gateway 10.10.1.254, host 10.10.1.253. DHCP do gateway oferece endereços na faixa observada 10.10.1.88–90, lease de 60s; não foi obtido o intervalo completo do pool. Os serviços de observabilidade e Tribunal usam IPs estáticos 200–219. CTID 220 confirmado livre pela API; candidato 10.10.1.220 não respondeu à sondagem ARP. Seleção planejada: CT220, endereço estático 10.10.1.220/24, verificar novamente antes do boot e registrar conflito se surgir. Não há reserva DHCP comprovada.

Template disponível: Debian 13.6 `local:vztmpl/debian-13-standard_13.6-1_amd64.tar.zst`. Planejado: `cells-prd-01`, não privilegiado, sem nesting, 4 vCPU com limite inicial 2 CPUs, 6144 MiB RAM, 512 MiB swap, rootfs 40 GiB no local-lvm, onboot, timezone America/Sao_Paulo. Limites não significam consumo medido ou reserva de RAM. Builds serializados; treino separado e limitado, sem compartilhar o processo de produção.

Publicação atual Malha: Cloudflare DNS e túneis dedicados; certificado de gestão existente permite listar os túneis. `cells.malha.app` estava ausente e nenhum túnel Cells existia. Criar túnel exclusivo local ao novo CT; não alterar ingress de terceiros. TLS/WSS terminam na Cloudflare, origem HTTP somente loopback. Firewall global PVE não será habilitado/alterado: aplicar nftables exclusivamente dentro do novo CT. PostgreSQL somente local; métricas privadas restritas à origem 10.10.1.200.

Observabilidade: Prometheus/Grafana nativos no CT200 (10.10.1.200), disco já próximo de 94%. Somente métricas agregadas de baixa cardinalidade e retenção existente, sem modelos, replays ou assets nesse CT. Integração exige backup de configuração, promtool e reload seguro; dashboard dedicado via API se houver credencial adequada. Não reiniciar Grafana de terceiros.

Backups Proxmox: job global Malha desativado; outro job ativo de terceiro, intocado. Não adicionar dumps de LXC ao filesystem quase cheio do host. Cells terá pg_dump/configuração próprios, retenção e restauração em banco isolado. Cópia externa depende de destino apropriado e será registrada como pendência se ausente.

GitHub: spy-exe autenticado com repo/workflow; `spy-exe/cells` não existe no reconhecimento. Criar público AGPL-3.0, sem reescrever histórico. Credenciais não serão copiadas para o Git.
