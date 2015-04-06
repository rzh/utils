#!/bin/bash

curl http://ftp.heanet.ie/mirrors/www.apache.org/dist/maven/maven-3/3.1.1/binaries/apache-maven-3.1.1-bin.tar.gz | sudo tar zx -C /usr/local
cd /usr/local
sudo ln -s apache-maven-* maven
echo "export M2_HOME=/usr/local/maven" | sudo tee -a /etc/profile.d/maven.sh
echo "export PATH=\${M2_HOME}/bin:\${PATH}" | sudo tee -a /etc/profile.d/maven.sh


