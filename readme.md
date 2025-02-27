![image](https://github.com/user-attachments/assets/9ab443b2-a35b-4e60-b5e9-f6ffda6527ce)


Why?

I wanted a learning project with JS, Websockets, and hosting a server that could be extensible as other learning opportunities arise.
I already have some ideas for kernel projects & firewalls for the server.
and, most importantly, resume fodder.

deployment from scratch:

    ansible-playbook ./playbooks/build.yml
    
    - make sure inventory reflects what exists in Digital Ocean & go to update DNS
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/configure.yml
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/deploy.yml -K

    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/observability.yml -K

new release:
    
    git tag v?.?.?
    
    git push origin v?.?.?

check out Grafana on localhost:

    // ssh tunnel remote host port 3000(Grafana) to localhost port 9999
    ssh -L :9999:<ip>:3000 bad-chess@<ip> -p 65332

    browse to localhost:9999

    
