Why?

I wanted a learning project with JS, Websockets, and hosting a server that could be extensible as other learning opportunities arise.
I already have some ideas for kernel projects & firewalls for the server.
and, most importantly, resume fodder.

deployment from scratch:

    ansible-playbook ./playbooks/build.yml
<<<<<<< HEAD
    - make sure inventory reflects what exists in Digital Ocean add IP to DNS for bad-chess
=======
    
    - make sure inventory reflects what exists in Digital Ocean
    
>>>>>>> 7098cf007858d351a7fb84d979f315740b4b4cc8
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/configure.yml
    
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/deploy.yml -kK

new release:
    
    git tag v?.?.?
    
    git push origin v?.?.?
