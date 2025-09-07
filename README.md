# 💸 Sistema de Processamento de Pagamentos 💸

Essa é minha participação na lendária [**Rinha de Backend 2025**](https://github.com/zanfranceschi/rinha-de-backend-2025)! 🥊
As demais branches contem estudos utilizando outras linguagens e modelagens utilizadas na participação de outros participantes.
Esta NAO é a solucao submetida. Ela foi trabalhada apos o fim do desafio.

## 🛠️ Tecnologias Utilizadas 👨‍💻

* **Go** - [Build simple, secure, scalable systems with Go](https://go.dev/)
* **FastHTTP** - [Fast HTTP for Go](https://github.com/valyala/fasthttp)
* **Redis** - [The Real-time Data Platform](https://redis.io/)
* **go-redis** - [Redis Go client](https://github.com/redis/go-redis)
* **HaProxy** - [The Reliable, High Perf. TCP/HTTP Load Balancer](https://www.haproxy.org/)

## 🚀 Como Rodar

### Suba tudo com Docker! 🐳

```bash
git clone https://github.com/macedot/rinha-2025-go
cd rinha-2025-go
docker compose up -d --build
```

## Execucao Local (Rinha Final)

### AMD Ryzen 9 5900X (24) @ 3.70 GHz

```json
{
  "total_liquido": 2017182.213585,
  "total_bruto": 1817532.5,
  "total_taxas": 150439.7615,
  "p99": {
    "valor": "0.45ms",
    "bonus": "21%",
  }
}
```

### Apple M1 (8) @ 3.20 GHz

```json
{
  "total_liquido": 1652712.678,
  "total_bruto": 1801773.2,
  "total_taxas": 149060.522,
  "p99": {
    "valor": "86.35ms",
    "bonus": "0%",
  }
}
```

## Repositório no GitHub

Curtiu? Dê uma olhada no [código fonte](https://github.com/macedot/rinha-2025-go) e deixe uma ⭐!

## ✨ Agradecimentos

Alguns autores que inspiraram este projeto (obrigado a todos!):

* [Alan Silva](https://github.com/alan-venv/rinha-de-backend-2025)
* [Anderson Gomes](https://github.com/andersongomes001/rinha-2025/)
* [Josiney Jr.](https://github.com/JosineyJr/rdb25_02)
* [Marchos Uchoa](https://git.uchoamp.dev/uchoamp/zig-pay)
* [Joyce Godinho Bosco](https://github.com/joycegodinho/rinha-2025)

Deixe uma ⭐ pra eles!
