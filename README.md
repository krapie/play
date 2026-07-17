# Play

Interactive sandbox experiments — small, self-contained tools built for fun. **Live:** [play.kevinprk.com](https://play.kevinprk.com)

## Experiments

| Experiment | Description |
|------------|-------------|
| **Reaction Time** | How fast can you click when the screen flips? |
| **AWS Latency** | Round-trip time from your browser to every AWS region |
| **Pick** | Add options, let randomness decide |
| **The Button** | One button. Everyone presses it. |
| **Hash Generator** | MD5, SHA-1, SHA-256, SHA-512 — all client-side |
| **Network Topologies** | Clos fat-tree vs RNG flat — interactive comparison of AWS data center fabric design |
| **Tracer** | Visualize the network path to any host, hop by hop on a map |

## Getting Started

```bash
docker build -t play .
docker run -p 8080:80 play   # http://localhost:8080
```
