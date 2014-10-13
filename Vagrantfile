# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.box_check_update = false

  config.vm.provision :shell, inline: <<EOS
    set -e

    # install docker
    curl -sSL https://get.docker.io/ubuntu/ | sudo sh

    # add vagrant to docker group
    gpasswd -a vagrant docker
EOS
end
