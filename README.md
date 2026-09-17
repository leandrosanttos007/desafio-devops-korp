# Projeto Korp — Desafio DevOps

Serviço HTTP em Go exposto via NGINX (proxy reverso), instrumentado com métricas
Prometheus e visualizado no Grafana. Todo o ambiente é provisionado por Ansible
com um único comando.

## Arquitetura

```
                    :80                    (rede bridge korp-net)
  cliente  ─────────────────▶  nginx  ──────────────▶  http-server-projeto-korp:8080
                                                              │  /metrics
                                                              ▼
                                                        prometheus:9090
                                                              │
                                                              ▼
                                                         grafana:3000
```

- **http-server-projeto-korp** (Go): endpoint `GET /projeto-korp` retorna
  `{"nome":"Projeto Korp","horario":"<UTC>"}`. Não expõe porta no host.
- **nginx**: proxy reverso, única porta publicada (`80`).
- **prometheus**: coleta as métricas em `/metrics`.
- **grafana**: dashboard provisionado automaticamente.

Os quatro serviços têm **healthcheck** no `docker-compose.yml`, e o `depends_on`
usa `condition: service_healthy`: o NGINX só sobe depois que o app está saudável
(evitando o erro `host not found in upstream`) e o Grafana só sobe depois do
Prometheus.

## Estrutura

```
projeto-korp/
├── app/                     # aplicação Go + Dockerfile (multi-stage)
├── nginx/                   # http-server-projeto-korp.conf (proxy reverso)
├── prometheus/              # prometheus.yml (scrape config)
├── grafana/                 # datasource + dashboard provisionados
├── docker-compose.yml       # app + nginx + prometheus + grafana
├── ansible/                 # automação (Parte 3)
└── README.md
```

## Como rodar

### Opção A — Ansible (provisiona tudo, inclusive o Docker)

```bash
cd ansible
ansible-playbook playbook.yml -K
```

Pré-requisito: `ansible` instalado (`sudo apt install -y ansible` ou `pipx install ansible`).
O `-K` pede a senha do `sudo` (o playbook usa `become` para instalar o Docker).
O playbook instala o Docker, cria a rede, faz o build, sobe os containers e,
ao final, faz um `curl` no serviço e exibe a resposta no console.

O playbook é **idempotente**: rodar de novo sem mudanças resulta em
`changed=0` (nada é recriado à toa).

### Opção B — Docker Compose (ambiente já com Docker)

```bash
docker network create korp-net      # a rede é external no compose
docker compose up -d --build
```

## Testar

```bash
curl http://localhost:80/projeto-korp
# {"nome":"Projeto Korp","horario":"2026-09-15T12:00:00Z"}
```

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin / admin) → dashboard "Projeto Korp - Observabilidade"

Estado de saúde dos containers:

```bash
docker compose ps          # coluna STATUS mostra "(healthy)"
```

## Métricas expostas (`/metrics`)

| Métrica | Tipo | Significado |
|---|---|---|
| `http_requests_total{method,path,status}` | counter | volume de requisições |
| `http_request_duration_seconds` | histogram | latência das requisições |
| `up{job="http-server-projeto-korp"}` | (gerada pelo Prometheus) | disponibilidade do serviço |

> O app também expõe `GET /health` (usado pelo healthcheck do container).
