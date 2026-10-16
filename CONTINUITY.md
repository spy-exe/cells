# Cells — missão ativa

## Pedido e escopo

O usuário mudou explicitamente de Tribunal para Cells em 16/10/2026 e autorizou implementação, infraestrutura pessoal, GitHub público, deploy e evolução. **Não parar após fundação:** objetivo desta execução é Classic multiplayer, bots táticos/AI Arena, treinamento real e avaliação, Teams/comunidade, CI/CD, carga, backup e relatório final. Ler SPEC.md e docs/INFRASTRUCTURE.md. Não acessar ambientes corporativos Opikode/SENAI/SENAC/CETIQT/Zeus/Hades/Atena/Poseidon. Não alterar outros projetos.

Tribunal foi deixado seguro na produção 0.6.0/8bcba81. Candidato 0.7.0 da branch feat/evidence-in-session (3c0b582) passou CI 38777176877; checkpoint 9201530 foi enviado ao GitHub, sem merge ou deploy. Não retomar Tribunal nesta missão.

## Infra real descoberta/provisionada

Host pessoal único pve, Proxmox9.2.4, Ryzen73700U,8threads,17.5GiBRAM,~6.6GiBdisponível e swap6.7/8GiBocupada. Filesystemhost só~2.5GiBlivre após deps Node; não instalar Go/PyTorch ou backups grandes nele. local-lvm156.87GiB,65.21%usado no início (~54.6GiBlivres). Nenhum recurso de terceiros limpo.

CT **220 cells-prd-01** provisionado e rodando, Debian13.6, não privilegiado, **sem nesting**,4vCPU/cpulimit2,6144MiBRAM,512MiBswap,40GiBthin local-lvm,onboot,America/Sao_Paulo. vmbr0,IPfixo **10.10.1.220/24**,gateway10.10.1.254,DNS1.1.1.1. CTID validado livre; IP não respondeu ARP antes. Pool DHCP completo não conhecido (ofertas88–90,lease60s); registro estático segue padrão de serviços200–219. Não há reserva DHCP comprovada.

systemd257 do template falhou em tmp.mount/run-lock.mount/dev-mqueue.mount; foram mascaradas somente essas montagens opcionais dentro do CT, diretórios tmp/lock com1777, resetfailed. Agora systemd running e nenhum failed. Não habilitar nesting por advertência genérica do PVE. Hostchrony sincronizado; relógio do LXC é o do host.

PostgreSQL17.11 nativo ativo, somente127.0.0.1/::1 e socketUnix. UsuárioLinuxcells,PGrolecells,DBcells criados; autenticaçãopeer local, sem senha fixa. Banco contém apenas cells_schema versão1 nesta fase. Go1.27.2 em /opt/cells/toolchains/go (SHAoficial ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 verificado). PATH de pctexec não inclui/usr/local/bin; usar caminho absoluto ou exportPATHincluindoGo/bin e/usr/sbin. gcc/libcdev instalados para teste-race. Não foi instalado Node no CT; buildweb usaNode24.18.0 já disponível no host/CI. Rootworkspace /root/proj/cells.

Firewall nftables exclusivo doCT ativo: tabela inet cells,inputDROP,lo/established/ICMPeND permitidos,métricas9108somente10.10.1.200. IPv6globalrecebido masentradafiltrada. SSH não permitido na rede,adminporpct. Não habilitar firewallglobalPVE (estádesabilitado).

Cloudflare: certificado de gestão já disponível em /root/outravida/.runtime/cert.pem foi usado sem copiá-lo. Túnel exclusivo **187205be-962b-4abf-b398-b1cadc09599f / cells-prd-01**,credencial própria ignorada em .runtime/tunnel.json e instalada /etc/cells/tunnel.json0600,cells-tunnelowner. Config /etc/cells/cloudflared.yml,origem127.0.0.1:8080,métricas127.0.0.1:2001. Cloudflared2026.10.1 instalado porSHAoficial3bfc539f45d2620c71debb7a86be0efa737ed5ccbad9eb85c787116900e5c586. DNS **cells.malha.app criado** via cloudflared route dns. Unit cells-tunnel habilitada/ativa. HTTPS já responde comTLSválido,porém **502 porque aplicação de produção ainda não foi ativada**. Não declarar fase0 publicada até/readyzok comSHA.

Units cells.service e cells-tunnel.service instaladas,env /etc/cells/cells.env0640root:cells. cells.service ainda sem /opt/cells/current/cells; validação do binário inexistente é esperada antes da primeira release. Túnel rodausercells-tunnel; app rodaráusercells,NoNewPrivileges,semcapabilities,MemoryMax1G,CPUQuota150%. Sem namespaces de filesystem que exigiriam nesting; código/releases rootowned. Não iniciar aplicação deprodução antes dos gates.

Prometheus/Grafana emCT20010.10.1.200 comdisco94%; configuração existente em/etc/prometheus/prometheus.yml,semfileSDobservado. Integrar scrape próprio combackup/promtool/reloadseguro; não reiniciar Grafana. Outros projetospreservados. Não adicionarvzdump ao disco quasecheio dohost. Treino deverá ocorrer fora doCTde jogo; CTID221foiconfirmadolivre mas **não provisionado**. Planejar laboratórioCells limitadoCPU/mem/disco e medircapacidadeantes. Nenhum treinamento foi executado.

## Código e verificação atual

Repo **https://github.com/spy-exe/cells**,público,AGPL-3.0,Gitspy-exeautenticado. main aindaébootstrapcdeba3b; trabalho em **feat/foundation**,commitsca52daf,b710bff,039af3d. PrimeirocommitcontinhatesteREDdehealth e nãoérelease. Não reescreverhistórico. Últimofeaturecommit/estadoverificarGit.

React19.3.1,Pixi8.22.0,Vite8.3.4,TS7.0.2,Vitest5.0.5,Playwright1.64.1,pinnedpackage-lock. Go módulo github.com/spy-exe/cells,pgx5.11.0. Frontendfoundation main.tsx: entrada originalcomorbsSVG/CSS decorativos,marcaoriginal,fontsSpaceGrotesk/Inter,controlesdemovimento,formulário **desabilitado enquanto status.modes=[]**. Não háengine/protocoloWS/bots/treino/Teams/comunidade ainda. Não apresentá-los como feitos. packages/protocol/status.tsvalidaAPI.

BackendGoapps/game-server/cmd/cells e internal/server: health/ready/status eassets,CSP/HSTS/HTTP→HTTPScomorigemfixa,PGping/readiness,cells_schema1. Estápreparadoparaexpansão. Errobindagorafalhacomexit1(paraRestart=on-failure);testeintegracaostartup.shreproduziuRED(exit0)epassouGREENcontraoutroservidorna8081.

GohealthTDDREDreal(Config/Newausentes)→GREEN. TestesGo-raceaprovados(duasfunçõesdeserver),vetaprovado;2unitáriosTS,lint/typecheck/build/auditdepssemvulnerabilidades,staticpremiumaudit0achados. E2Elocal **2passaram desktop/mobile**,axe0,layouts360×800/390×844/412×915/768×1024/1440×900,preferênciapersistente. Chromiumhost/ root/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome. Comandos: npmtest/npmruntypecheck/npmrunbuild/npmrunlint. Go via CTouCI.

QArodando **cells-qa.service**,usercells,8081privada,binário/tmp/cells-foundation,web/opt/cells/build/dist/web. Regra nft temporária cells-qapermite8081somentehost10.10.1.253. **Parar QA e remover somente essa regra antes do deployhelper**,pois candidato usa8081loopback. EssaQAfoiisoladadoprod,semusuáriosreais.

Buildhelper infra/deploy/sync-build.sh transportafontessem.git/node_modules/.runtime/.envpara/opt/cells/build. Go cache/toolchains noCT. AtençãoPath: /usr/sbin/runuser ou PATHcompleto. Rootnão temGo. Nodehost jáinstalado24.18,nãofoialterado. FontesGoformatadasnoCTe trazidasdevoltaporpctpull. RodarGo-raceaoalterarmotor.

GitHubCI .github/workflows/ci.yml: hostedUbuntu,Postgres17containerCItrustsóloopback,npmci,Goformat/vet/race,TS/lint/unit/audit,build,Playwright,artefatomanifest/SHA256. Publishjobcontentswriteapenasmainpush,GitHubreleaseprereleasebuild-<sha>depoisverify;nenhumrunnerpúbliconohomelab.

CI38905636089falhouantesdostestesporquotinghealthcmddoservice;corrigidoparacmdúnicopg_isready. CI38905931698passouGo/TS,masfalhoubuildporpublicdirgeradolicençasnãotrackeado;corrigidomkdirrecurseefaviconoriginaltrackeado. **CIatual039af3dverificar**,nãoassumiraprovado. Scriptsnotices preservalicençasNPMempublic/licenses.txt(ignorado,geradobuild). scripts/go-notices.pyagorafoiescritoparapreservarlicençasdosmódulosGonoartefato; **alteraçãoaindapodeestarsemcommit**,verGit. Secretpattern gate scripts/check-secrets.pysemsecretdetectado.

## Próximo passo imediato

1. Terminar/testarcommitdoshelpersdeCDelicençasGo; conferirCIatual, corrigirfalhasreais, não fazerdeploycomgatefalhando.
2. CDroothelper infra/deploy/pull-release.py estáimplementadomasaindanãoinstalado/executado: consultaCIbem-sucedidomain/push,releasebuildsha,checksum,tarsemtraversal/links,ELF/manifest;backupPGantesdecandidato;candidateusercellsem8081;trocaatomiccurrent/readinesslocal/HTTPS;rollbacksemreverterDB;--rollbackcriahold,--resumeretoma. Revisar/testarsegurança/transições antesdeinstalarrootownedem/usr/local/libexec/cells. Unitsdeploy.service/timer5minaindaapenasfontes. APIpúblicasemtokenGitHub,semcredenciaisinfraemActions. Fixardocs-onlypoll404futurosepublicaçãoforfiltradapordocs(atualCIexecutatodosospushes).
3. QuandofeatureCIverde,fast-forwardmainsemanularhistórico,push,esperarverify+publishdomain. Pararcells-qa,removerregra8081ownQA,instalarhelper/units. PrimeirodeployviahelperdevebaixarartefatoCIeativar/readyzversão0.0.1SHAexato;validarHTTPS/redirectexternos,smoke,PGbackup/restoreisolado. Nãoesperarprodparaimplementarremainingarquiteturaexcessiva: objetivoévidaClassicjogávelrapidamente.
4. **Continuar imediatamente Phase1**: engineGoindependentedarede,TDDdeterminístico(seedserializável),spatialgrid,protocoloWSbinário,frontendPixi,guestHTTPcookieHttpOnly/sessionPG,split/eject/virus/merge,múltiplascélulas,respawn/spectator/minimap/leaderboard/reconnect. Semfakefeatures. Depoisbotutilityprofile5+AIarena,treinoheadlessPPOreal100episseedscomparativos(heurísticanaprodsemodelopior),Teams/community,load25/50/100,metrics/backups/rollbackeE2Epúblico2clientes.

Não entregar apenasessa fundação como missão concluída. O usuário autorizou a execução completa por fases. Atualizaçõesemportuguês≤60s,semperguntasdesnecessárias. Nãospawnsubagentessemrequisiçãoexplícitaouinstruçãodeskillaplicável. Nenhumagentenovoficoutrabalhando;social_reproduction_testsélegadoOUTRAVIDAconcluído e irrelevante.
