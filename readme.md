Why?

I wanted a learning project with JS, Websockets, and hosting a server that could be extensible as other learning opportunities arise.
I already have some ideas for kernel projects & firewalls for the server.
and, most importantly, resume fodder.

deployment from scratch:

    ansible-playbook ./playbooks/build.yml
    
    - make sure inventory reflects what exists in Digital Ocean & go to update DNS
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/configure.yml
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/deploy.yml -kK

new release:
    
    git tag v?.?.?
    
    git push origin v?.?.?
