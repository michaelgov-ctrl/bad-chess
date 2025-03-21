# Welcome to bad-chess


![image](https://github.com/user-attachments/assets/810f5ed2-caff-4b92-926c-eb2032217a71)


Please feel free to play an instance of [Stockfish](https://github.com/official-stockfish/Stockfish) or try to find a match in matchmaking. The Stockfish engine is licensed under GPL-3.0


![let-me-in](https://github.com/user-attachments/assets/506cabe4-d0ae-4dc6-922b-b62d7e33d7d0)



Authentication is needed under a not-so-secret-key to protect the websocket used for matches from being abused(for now).
The key is WelcomeToBadChess


![image](https://github.com/user-attachments/assets/ff7f0197-8e42-4530-bc65-1337c8f8c305)


## Why?

I set out to build a chess-based website as a learning project to focus on JavaScript, WebSockets, and was extensible. I wanted to work with 'real-time' communication and WebSockets provided a new(to me) opportunity to explore that. With the core infrastructure in place, I have ideas for additional projects related to security, such as implementing custom firewall rules and experimenting with kernel-level optimizations. This project gives me a platform to explore those ideas in a practical setting.

## binary help

    Usage:
      myapp [flags]
    
    Flags:
      -port int
            API server port (default 8080)
      -loki-port int
            Port of local Loki instance to log to (optional)
      -log-level string
            Logging level (trace|debug|info|warning|error) (default "error")
      -cert string
            File containing certificate for TLS (optional)
      -key string
            File containing key for TLS (optional)
      -cors-trusted-origins string
            Trusted CORS origins (space-separated)

## deployment from scratch:

    ansible-playbook ./playbooks/build.yml
    
    - make sure inventory reflects what exists in Digital Ocean & go to update DNS
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/configure.yml
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/deploy.yml -K

    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/observability.yml -K

## new release:
    
    git tag v?.?.?
    
    git push origin v?.?.?

## check out Grafana on localhost:

    // ssh tunnel remote host port 3000(Grafana) to localhost port 9999
    ssh -L :9999:<ip>:3000 bad-chess@<ip> -p 65332

    browse to localhost:9999

## Observability:

The basic observability stack is Alloy, Prometheus, Loki, & Grafana.
Prometheus scrapes the bad-chess server that has been instrumented for Prometheus metrics and Node Exporter.
Alloy scrapes logs for fail2ban & Caddy and ships them to Loki. The bad-chess server ships its own logs to Loki.
To visualize these metrics & logs at a glance there are three Grafana Dashboards: bad-chess, Node Exporter Full, and Logs.


![image](https://github.com/user-attachments/assets/6221b244-23bd-43e7-9e31-c32b410c405b)


### bad-chess
![image](https://github.com/user-attachments/assets/37071088-ba39-468b-9562-562a5bc8451b)


### Node Exporter Full
![image](https://github.com/user-attachments/assets/27b1a2c5-ddb0-49c4-a756-0e307f015dd5)


### Logs
![image](https://github.com/user-attachments/assets/e9e7adad-c6a4-422f-8e3e-2143596d4091)


