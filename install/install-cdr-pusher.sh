#!/bin/bash
#
# The primary maintainer of this project is Areski Belaid <areski@gmail.com>
#
#
# To download and run the script on your server :
# cd /usr/src/ ; rm install-cdr-pusher.sh ; wget --no-check-certificate https://raw.github.com/areski/cdr-pusher/master/install/install-cdr-pusher.sh ; chmod +x install-cdr-pusher.sh ; ./install-cdr-pusher.sh
#


#Set branch to install develop / master
BRANCH="master"
DATETIME=$(date +"%Y%m%d%H%M%S")
INSTALL_DIR='/opt/app'


#Django bug https://code.djangoproject.com/ticket/16017
export LANG="en_US.UTF-8"
SCRIPT_NOTICE="This script is only intended to run on Debian, Ubuntu or Redhat"

# Identify Linux Distribution type
func_identify_os() {
    if [ -f /etc/debian_version ] ; then
        DIST='DEBIAN'
    elif [ -f /etc/redhat-release ] ; then
        DIST='REDHAT'
    else
        echo $SCRIPT_NOTICE
        exit 1
    fi
}


#Function to install Dependencies
func_install_dependencies(){

    #python setup tools
    echo "Install Dependencies and python modules..."

    case $DIST in
        'DEBIAN')
            apt-get -y install golang
            apt-get -y install mercurial git bzr
            apt-get -y install supervisor
        ;;
        'CENTOS')
            yum -y groupinstall "Development Tools"
            yum -y install golang
            yum -y install mercurial git bzr
            yum -y install supervisor
        ;;
    esac
}


#function to get the source and install
func_install(){
    echo "Install CDR-Pusher..."
    mkdir /opt/app
    cd /opt/app

    git clone -b $BRANCH git://github.com/areski/cdr-pusher.git
    cd cdr-pusher

    #Install cdr-pusher
    export GOPATH=`pwd`
    make build
    make logdir

    #Install supervisor
    make install-supervisor-conf

    /etc/init.d/supervisor stop
    sleep 2
    /etc/init.d/supervisor start
}

#Configure Log dir and logrotate
func_prepare_logrotate() {
    echo "Install Logrotate..."
    rm /etc/logrotate.d/cdr_pusher
    touch /etc/logrotate.d/cdr_pusher
    echo '
    /var/log/cdr-pusher/*.log {
        daily
        rotate 10
        size = 50M
        missingok
        compress
    }
    '  >> /etc/logrotate.d/cdr_pusher

    logrotate /etc/logrotate.d/cdr_pusher
}


#========== Start Installation ==========

#Identify the OS
func_identify_os

#Install deps
func_install_dependencies

#Install conf logrotate
func_prepare_logrotate

#Install
func_install
