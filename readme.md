Why?
I wanted a learning project with JS & Websockets that could be extensible as other learning opportunities arise.
I already have some ideas for kernel projects for the server.
and, most importantly, resume fodder.

deployment from scratch:
    ansible-playbook ./playbooks/build.yml
    - make sure inventory reflects what exists in Digital Ocean add IP to DNS for bad-chess
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/configure.yml
    ansible-playbook -i ./playbooks/inventory/inventory ./playbooks/deploy.yml -kK

new release:
    git tag v?.?.?
    git push origin v?.?.?