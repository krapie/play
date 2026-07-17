# Play

Interactive sandbox experiments — small, self-contained tools built for fun and curiosity. Each experiment runs entirely in the browser with no backend dependency (except AWS Latency and Tracer which hit external endpoints). **Live:** [play.kevinprk.com](https://play.kevinprk.com)

## Getting Started

```bash
docker build -t play .
docker run -p 8080:80 play   # http://localhost:8080
```

## Experiments

- **Reaction Time** — measure how fast you can click when the screen changes color; shows your average over multiple rounds
- **AWS Latency** — pings all AWS regions from your browser and ranks them by round-trip time; useful for picking a deployment region
- **Pick** — add a list of options and let the app pick one at random; great for indecisive moments
- **The Button** — a single shared button; every visitor can press it and see the global count go up
- **Hash Generator** — MD5, SHA-1, SHA-256, and SHA-512 hashing, all computed client-side with no data sent anywhere
- **Network Topologies** — interactive visual comparison of Clos fat-tree vs. RNG flat topology used in AWS data center fabric design
- **Tracer** — enter any hostname and visualize the network path hop by hop on a world map
