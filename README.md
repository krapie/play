# play

Interactive sandbox experiments — small, self-contained tools I built for no particular reason.

**Live:** https://play.kevinprk.com

## Experiments

| Experiment | Description |
|---|---|
| **Reaction Time** | How fast can you click when the screen flips? |
| **AWS Latency** | Round-trip time from your browser to every AWS region |
| **Pick** | Add options, let randomness decide |
| **The Button** | One button. Everyone presses it. |
| **Hash Generator** | MD5, SHA-1, SHA-256, SHA-512 — all client-side |
| **Tracer** | Visualize the network path to any host, hop by hop on a map |

## Stack

| Layer | Tech |
|---|---|
| Frontend | Vanilla HTML / CSS / JS |
| Backend API | Go + chi + SQLite |
| Serving | Nginx (static) |
| Deploy | Docker + Kubernetes (ArgoCD) |
| CI | GitHub Actions |

## Project Structure

```
play/
├── web/
│   ├── index.html          # experiment listing
│   ├── colors_and_type.css # design tokens (--kp-*)
│   ├── manifest.json
│   ├── button/             # The Button
│   ├── hash/               # Hash Generator
│   ├── latency/            # AWS Latency
│   ├── pick/               # Pick
│   ├── reaction/           # Reaction Time
│   └── tracer/             # Tracer
├── api/
│   ├── main.go
│   ├── handlers/           # HTTP handlers
│   └── db/                 # SQLite
├── nginx.conf
└── Dockerfile
```

## Local Setup



## CI/CD

Push to `main` → GitHub Actions builds `krapi0314/play:<sha>` and pushes to Docker Hub → updates `k8s/play/deployment.yaml` in [krapie/homeserver](https://github.com/krapie/homeserver) → ArgoCD syncs to the cluster.
