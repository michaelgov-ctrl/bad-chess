# Welcome to bad-chess


![image](https://github.com/user-attachments/assets/f72eabd9-30fb-439c-a907-fc16cac08fd9)


At the moment only matchmaking is available, so the likelihood of finding an opponent is low....
Games against an engine is the next goal.


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

    
