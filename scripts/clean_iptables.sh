#!/bin/bash

# Function to delete MINIK8S chains in the specified order
delete_minik8s_chains() {
    # List all MINIK8S chains in the nat table and grep for them
    chains=$(sudo iptables -L -n -t nat | grep '^Chain MINIK8S-' | awk '{print $2}')

    # Delete SVC chains
    svc_chains=$(echo "$chains" | grep '^SVC')
    for chain in $svc_chains; do
        echo "Deleting iptables chain $chain ..."
        sudo iptables -t nat -F $chain
        sudo iptables -t nat -X $chain
    done

    # Delete SEP chains
    sep_chains=$(echo "$chains" | grep '^SEP')
    for chain in $sep_chains; do
        echo "Deleting iptables chain $chain ..."
        sudo iptables -t nat -F $chain
        sudo iptables -t nat -X $chain
    done

    # Delete MARK-MASQ chain
    mark_masq_chain=$(echo "$chains" | grep '^MARK-MASQ')
    echo "Deleting iptables chain $mark_masq_chain ..."
    sudo iptables -t nat -F $mark_masq_chain
    sudo iptables -t nat -X $mark_masq_chain

    # Delete POSTROUTING chain
    postrouting_chain=$(echo "$chains" | grep '^POSTROUTING')
    echo "Deleting iptables chain $postrouting_chain ..."
    sudo iptables -t nat -F $postrouting_chain
    sudo iptables -t nat -X $postrouting_chain

    # Delete SERVICES chain
    services_chain=$(echo "$chains" | grep '^SERVICES')
    echo "Deleting iptables chain $services_chain ..."
    sudo iptables -t nat -F $services_chain
    sudo iptables -t nat -X $services_chain
    sudo systemctl restart docker

}

# Clean MINIK8S chains in the specified order
echo "Cleaning MINIK8S iptable chains ..."
delete_minik8s_chains

echo "MINIK8S iptable chains cleaned."
sudo iptables -L -n -t nat | grep '^Chain MINIK8S-'